package dianzong

import (
	"agent/entity/consts"
	"agent/logic/collector/device/model"
	"agent/utils"
	"agent/utils/osal"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"trpc.group/trpc-go/trpc-go/log"
)

// 常量：依据协议定义
const (
	SOI = 0x7E
	EOI = 0x0D
)

// BuildSimpleCommand 构建无/有 INFO 的标准命令
// - INFO 以“原始字节”传入，本函数会转为 ASCII HEX 并计算 LENGTH/CHKSUM
func BuildSimpleCommand(ver, addr, cid1, cid2 byte, info []byte) []byte {
	// 用 hexAppend 拼 INFO（大写）
	var infoASCII []byte
	if len(info) > 0 {
		infoASCII = make([]byte, 0, len(info)*2)
		for _, b := range info {
			hexAppend(&infoASCII, b)
		}
	}

	lenID := uint16(len(infoASCII))
	length := packLength(lenID)
	lengthASCII := []byte(fmt.Sprintf("%04X", length)) // 大写

	checkPart := make([]byte, 0, 1+1+1+1+4+len(infoASCII))
	hexAppend(&checkPart, ver)
	hexAppend(&checkPart, addr)
	hexAppend(&checkPart, cid1)
	hexAppend(&checkPart, cid2)
	checkPart = append(checkPart, lengthASCII...)
	checkPart = append(checkPart, infoASCII...)

	chk := checksumASCII(checkPart)
	chkASCII := []byte(fmt.Sprintf("%04X", chk))

	out := []byte{SOI}
	out = append(out, checkPart...)
	out = append(out, chkASCII...)
	out = append(out, EOI)
	return out
}

// BuildRequestFromPacket 依据 CollectProtocolPacket.Command 构建请求帧
// 约定格式如下：
//   - "2A44"            → CID1=0x2A, CID2=0x44, INFO=[]
//   - "2AE70100"        → CID1=0x2A, CID2=0xE7, INFO=[0x01,0x00]
//   - 大小写、带/不带 0x 前缀、含空格/下划线/连字符均可，被 parseHexString 归一化
func BuildRequestFromPacket(ver, addr byte, p *model.CollectProtocolPacket) ([]byte, func(*DecodedFrame,
	*model.CollectProtocolPacket) error, error) {
	if p == nil || strings.TrimSpace(p.Command) == "" {
		return nil, nil, errors.New("BuildRequestFromPacket: empty packet or Command")
	}

	cmdBytes, err := parseHexString(p.Command)
	if err != nil {
		return nil, nil, fmt.Errorf("BuildRequestFromPacket: invalid Command hex: %w", err)
	}
	if len(cmdBytes) < 2 {
		// 明确不支持“只给 CID2”的 1 字节写法
		return nil, nil, errors.New("BuildRequestFromPacket: Command must contain CID1 and CID2 (>=2 bytes)")
	}

	// 拆出 CID1/CID2，剩余全部作为 INFO
	cid1 := cmdBytes[0]
	cid2 := cmdBytes[1]
	var info []byte
	if len(cmdBytes) > 2 {
		info = append([]byte{}, cmdBytes[2:]...)
	}

	req := BuildSimpleCommand(ver, addr, cid1, cid2, info)
	return req, decideParsePlan(cid2), nil
}

// parseHexString 将 "2A44" / "0x2a44" / "2a 44" / "2A-E7-0100" 等解析为字节序列
func parseHexString(s string) ([]byte, error) {
	if s == "" {
		return nil, nil
	}
	clean := strings.ToLower(strings.TrimSpace(s))
	clean = strings.ReplaceAll(clean, "0x", "")
	clean = strings.ReplaceAll(clean, " ", "")
	clean = strings.ReplaceAll(clean, "_", "")
	clean = strings.ReplaceAll(clean, "-", "")
	if len(clean)%2 != 0 {
		return nil, fmt.Errorf("odd hex length: %d", len(clean))
	}
	dst := make([]byte, len(clean)/2)
	_, err := hex.Decode(dst, []byte(clean))
	if err != nil {
		return nil, err
	}
	return dst, nil
}

// BuildControlInfoFromPacket 将控制包映射为 (COMMAND_TYPE, COMMAND_ID, EXTRA)
// - 示例：val = "power:on" / "battery_test:start" 等
func BuildControlInfoFromPacket(_ *model.ControlProtocolPacket, val string) (byte, byte, []byte, error) {
	switch strings.ToLower(val) {
	case "battery_test:start":
		return 0x10, 0x01, nil, nil // 6.10.1
	case "battery_test:stop":
		return 0x10, 0x02, nil, nil
	case "power:on":
		return 0x20, 0x01, nil, nil
	case "power:off_inverter":
		return 0x20, 0x02, nil, nil
	case "power:off_now":
		return 0x20, 0x03, nil, nil
	case "open:delay":
		return 0x20, 0x04, nil, nil
	case "close:delay":
		return 0x20, 0x05, nil, nil
	case "close:delay_cancel":
		return 0x20, 0x06, nil, nil
	case "battery_maintain:start":
		return 0x30, 0x01, nil, nil
	case "battery_maintain:stop":
		return 0x30, 0x02, nil, nil
	default:
		return 0, 0, nil, fmt.Errorf("unsupported control val: %s", val)
	}
}

// ——— 解码/校验 ——— //

type ParsedBag map[string]any

type DecodedFrame struct {
	VER   byte
	ADR   byte
	RTN   byte // 对应 CID2 位置的返回码/或 CID2 回显
	CID1  byte
	CID2  byte
	INFO  []byte // ASCII 原样（含 DATAFLAG + DATAI/DATAF等），上层再解码
	RAW   []byte // 完整帧
	OK    bool
	Error string

	Parsed ParsedBag //规范化后的结构化数据
}

// —— 公共辅助 —— //

// 未使用：哪些 CID2 的响应首字节是 DATAFLAG（剥旗后 reg 要整体 -1）
var framesWithDataFlag = map[byte]bool{
	0x41: true, 0x42: true,
	0x43: true, 0x44: true, 0x89: true, 0x8A: true,
	0xE0: true, 0xE1: true, 0xE2: true, 0xE3: true, 0xE4: true,
}

func parseUnified(df *DecodedFrame, p *model.CollectProtocolPacket) error {
	if p == nil {
		return fmt.Errorf("nil packet")
	}
	reqCID2, err := cid2FromCommand(p.Command)
	if err != nil {
		return err
	}
	raw, ok := df.RawPayload()
	if !ok {
		return fmt.Errorf("raw payload decode failed")
	}
	var flag byte
	if framesWithDataFlag[reqCID2] && len(raw) > 0 {
		flag = raw[0]
	}
	if df.Parsed == nil {
		df.Parsed = ParsedBag{}
	}
	df.Parsed["DATAFLAG"] = flag
	df.Parsed["PAYLOAD"] = raw
	return mapPointsByReg(df, p, reqCID2)
}

// 仅保留 0–9 / a–f / A–F
// 把 INFO (ASCII) 转原始字节 —— 兼容空格/占位;见协议：不支持的数据用空格
func (df *DecodedFrame) infoRaw() ([]byte, error) {

	b := asciiHexDecodePairs(df.INFO)
	if len(b) == 0 {
		return nil, fmt.Errorf("empty INFO after pairwise decode, raw=%q", string(df.INFO))
	}
	return b, nil
}

func (df *DecodedFrame) payloadAfterFlag(_ byte) (byte, []byte, error) {
	raw, err := df.infoRaw()
	if err != nil {
		return 0, nil, err
	}
	if len(raw) > 0 {
		return raw[0], raw, nil
	}
	return 0, raw, nil
}

func (df *DecodedFrame) payloadAfterFlagAuto(_ byte) (byte, []byte, error) {
	raw, ok := df.RawPayload()
	if !ok {
		return 0, nil, fmt.Errorf("raw payload decode failed")
	}
	if len(raw) > 0 {
		return raw[0], raw, nil
	}
	return 0, raw, nil
}

// 工具：大小端读取
func u16(b []byte, order binary.ByteOrder) (uint16, error) {
	if len(b) < 2 {
		return 0, errors.New("u16: not enough bytes")
	}
	return order.Uint16(b[:2]), nil
}
func u32(b []byte, order binary.ByteOrder) (uint32, error) {
	if len(b) < 4 {
		return 0, errors.New("u32: not enough bytes")
	}
	return order.Uint32(b[:4]), nil
}

func ParseResponse(frame []byte) (*DecodedFrame, error) {
	if len(frame) < 1 || frame[0] != SOI || frame[len(frame)-1] != EOI {
		return nil, errors.New("bad frame boundary")
	}
	ascii := frame[1 : len(frame)-1]

	if len(ascii) < 2*4+4 {
		return nil, errors.New("frame too short")
	}

	ver, err := readHexByte(ascii[0:2])
	if err != nil {
		return nil, fmt.Errorf("VER: %w", err)
	}
	adr, err := readHexByte(ascii[2:4])
	if err != nil {
		return nil, fmt.Errorf("ADR: %w", err)
	}
	cid1, err := readHexByte(ascii[4:6])
	if err != nil {
		return nil, fmt.Errorf("CID1: %w", err)
	}
	rtn, err := readHexByte(ascii[6:8])
	if err != nil {
		return nil, fmt.Errorf("RTN: %w", err)
	}

	hi, err := readHexByte(ascii[8:10])
	if err != nil {
		return nil, fmt.Errorf("LENGTH.hi: %w", err)
	}
	lo, err := readHexByte(ascii[10:12])
	if err != nil {
		return nil, fmt.Errorf("LENGTH.lo: %w", err)
	}
	length := (uint16(hi) << 8) | uint16(lo)
	lchk := byte(length >> 12)
	lenID := length & 0x0FFF

	head := ascii[:12]
	if !checkLENGTH(lchk, lenID) {
		return nil, errors.New("LENGTH LCHKSUM invalid")
	}

	expect := 12 + int(lenID) + 4
	// Tolerant: compare pair counts and allow one trailing nibble (off-by-one)
	if pairs, expPairs := len(ascii)/2, expect/2; pairs != expPairs {
		return nil, fmt.Errorf("invalid total ascii pairs=%d expect=%d (ascii len=%d expect=%d)", pairs, expPairs, len(ascii), expect)
	}
	info := ascii[12 : 12+lenID]
	chkASCII := ascii[12+lenID : 12+lenID+4]
	chk, err := readHexU16(chkASCII)
	if err != nil {
		return nil, fmt.Errorf("CHKSUM field: %w", err)
	}

	sumPart := append(append([]byte{}, head...), info...)
	if checksumASCII(sumPart) != chk {
		return nil, errors.New("CHKSUM invalid")
	}

	df := &DecodedFrame{VER: ver, ADR: adr, RTN: rtn, CID1: cid1, CID2: 0,
		INFO: append([]byte{}, info...), RAW: frame, OK: rtn == 0x00}
	return df, nil
}

// 新增
func readHexByte(ascii []byte) (byte, error) {
	if len(ascii) != 2 {
		return 0, fmt.Errorf("bad hex byte len=%d", len(ascii))
	}
	b, err := hex.DecodeString(string(ascii))
	if err != nil || len(b) == 0 {
		return 0, fmt.Errorf("bad hex byte: %q", string(ascii))
	}
	return b[0], nil
}

func readHexU16(ascii []byte) (uint16, error) {
	if len(ascii) != 4 {
		return 0, fmt.Errorf("bad hex u16 len=%d", len(ascii))
	}
	b, err := hex.DecodeString(string(ascii))
	if err != nil || len(b) < 2 {
		return 0, fmt.Errorf("bad hex u16: %q", string(ascii))
	}
	return (uint16(b[0]) << 8) | uint16(b[1]), nil
}

// ———— 工具 ———— //

func hexAppend(dst *[]byte, b byte) {
	*dst = append(*dst, []byte(fmt.Sprintf("%02X", b))...)
}

// LENGTH 打包：LENGTH = [LCHKSUM(4bits)][LENID(12bits)]
func packLength(lenID uint16) uint16 {
	sum := ((lenID >> 8) & 0x0F) + ((lenID >> 4) & 0x0F) + (lenID & 0x0F)
	lchk := byte((-int(sum)) & 0x0F)
	return (uint16(lchk) << 12) | (lenID & 0x0FFF)
}

func checkLENGTH(lchk byte, lenID uint16) bool {
	sum := ((lenID >> 8) & 0x0F) + ((lenID >> 4) & 0x0F) + (lenID & 0x0F)
	expect := byte((-int(sum)) & 0x0F)
	return lchk == expect
}

func checksumASCII(ascii []byte) uint16 {
	var s uint32
	for _, b := range ascii {
		s += uint32(b)
	}
	// 取反加 1（等价于 0 - s）
	return uint16(^s + 1)
}

// —— 可选：把命令名转 CID2 —— //
func nameToCID2AndInfo(name string, _ *model.CollectProtocolPacket) (byte, []byte, error) {
	switch strings.ToUpper(name) {
	case "AIM_42", "ANALOG_42":
		return 0x42, nil, nil
	case "INPUT_E0":
		return 0xE0, nil, nil
	case "OUTPUT_E1":
		return 0xE1, nil, nil
	case "SYS_OUTPUT_E2":
		return 0xE2, nil, nil
	case "BAT1_E3":
		return 0xE3, nil, nil
	case "BAT2_E4":
		return 0xE4, nil, nil
	case "DI_43":
		return 0x43, nil, nil
	case "ALARM_44":
		return 0x44, nil, nil
	case "PARAM_47":
		return 0x47, nil, nil
	case "VER_4F":
		return 0x4F, nil, nil
	case "ADDR_50":
		return 0x50, nil, nil
	case "VENDOR_51":
		return 0x51, nil, nil
	case "FAULT_89":
		return 0x89, nil, nil
	case "EVENT_8A":
		return 0x8A, nil, nil
	default:
		return 0, nil, fmt.Errorf("unknown command name: %s", name)
	}
}

// 根据 CID2 决定解析策略
func decideParsePlan(cid2 byte) func(*DecodedFrame, *model.CollectProtocolPacket) error {
	return parseUnified
	// 不支持特殊命令了
	//switch cid2 {
	//case 0x42:
	//	return parseDATAI_42
	//case 0xE0, 0xE1, 0xE2, 0xE3, 0xE4:
	//	return parseDATAF_Ex
	//case 0x43, 0x44, 0x89, 0x8A:
	//	return parseStates
	//default:
	//	return nil
	//}
}

// 6.2.1 标准 0x42
func parseDATAI_42(df *DecodedFrame, p *model.CollectProtocolPacket) error {
	flag, payload, err := df.payloadAfterFlag(0x42)
	if err != nil {
		return err
	}
	if len(payload)%2 != 0 {
		return fmt.Errorf("42 payload len must be even, got %d", len(payload))
	}
	n := len(payload) / 2
	out := make([]uint16, n)
	for i := 0; i < n; i++ {
		out[i] = binary.BigEndian.Uint16(payload[i*2 : i*2+2])
	}
	if df.Parsed == nil {
		df.Parsed = ParsedBag{}
	}
	df.Parsed["DATAFLAG"] = flag
	df.Parsed["DATAI16"] = out

	// 把值按配置回填到测点（命令 0x42）
	return mapPointsByReg(df, p, 0x42)
}

// 6.2.2~6.2.6 自定义模拟帧 E0/E1/E2/E3/E4：文档称 DATAF
func parseDATAF_Ex(df *DecodedFrame, p *model.CollectProtocolPacket) error {
	_, payload, err := df.payloadAfterFlag(0xE0) // E0~E4 都有 DATAFLAG，这里共用 0xE0 走位移
	if err != nil {
		return err
	}
	if df.Parsed == nil {
		df.Parsed = ParsedBag{}
	}
	df.Parsed["DATAFLAG"] = payloadFlagIfAny(payload)
	df.Parsed["DATAF_RAW"] = payload

	// 实际请求的 CID2 取自 packet.Command（可能是 E0/E1/E2/E3/E4 中之一）
	reqCID2, _ := cid2FromCommand(p.Command)
	return mapPointsByReg(df, p, reqCID2)
}

// 0x43/0x44/0x89/0x8A：RUNSTATE/WARNSTATE/FAULT/EVENT，都是 1B=00/ F0 类型的状态表
func parseStates(df *DecodedFrame, p *model.CollectProtocolPacket) error {
	_, payload, err := df.payloadAfterFlag(0x43) // 0x43/0x44/0x89/0x8A 结构一致
	if err != nil {
		return err
	}
	if df.Parsed == nil {
		df.Parsed = ParsedBag{}
	}
	df.Parsed["STATE_BYTES"] = payload

	// 把值按配置回填到测点
	reqCID2, _ := cid2FromCommand(p.Command)
	return mapPointsByReg(df, p, reqCID2)
}

func cid2FromCommand(cmd string) (byte, error) {
	bs, err := parseHexString(cmd)
	if err != nil {
		return 0, err
	}
	if len(bs) < 2 {
		return 0, fmt.Errorf("cid2FromCommand: need at least 2 bytes")
	}
	return bs[1], nil
}

func mapPointsByReg(df *DecodedFrame, packet *model.CollectProtocolPacket, reqCID2 byte) error {
	if packet == nil || len(packet.Points) == 0 {
		return nil
	}
	now := utils.GetNowUTCTimeStamp()

	for _, pt := range packet.Points {
		if pt == nil || pt.Attr.ValParser == nil {
			continue
		}
		vp, ok := pt.Attr.ValParser.(*ValueParser)
		if !ok {
			continue
		}

		// 2) 抽取并写值：默认 big endian/uint8（状态帧 0x43/0x44/0x89/0x8A 为 1 字节 00/F0）
		reg := strings.TrimSpace(vp.Addr)
		if reg == "" {
			continue
		}

		effectiveReg := reg

		dt := vp.DataType
		if dt == "" {
			if reqCID2 == 0x43 || reqCID2 == 0x44 || reqCID2 == 0x89 || reqCID2 == 0x8A {
				dt = "uint8"
			} else if reqCID2 == 0x42 {
				dt = "uint16"
			} else {
				dt = "uint16" // DATAF 默认 2B 定点为主（个别为 4B，按配置覆盖）
			}
		}
		endian := vp.ByteOrder
		if endian == "" {
			endian = "big"
		}

		val, err := ReadByReg(df, reqCID2, effectiveReg, dt, endian)
		if err != nil {
			// 某些寄存器越界/暂不支持，跳过该点（仅调试级别打印，避免大量日志）
			log.Debugf("ReadByReg skip point=%s reg=%s err:%s", pt.Attr.ID, effectiveReg, err)
			continue
		}

		// ext function: not..
		val = utils.Unary(vp.Extend, val)

		// —— 实际写回
		pt.RtVal.Pv = osal.NewVariantWithValue(val)

		pt.RtVal.Tms = now
		pt.RtVal.Qua = consts.QualityOk
	}
	return nil
}

// 仅为占位：把第一个字节视作 flag；当我们已经从 df.payloadAfterFlag 去掉 flag 时，这里返回 0
func payloadFlagIfAny(_ []byte) byte { return 0 }

// RawPayload : 无论有无 DATAFLAG，都把 INFO（ASCII）按“成对解码”转成原始字节；
// 空格 0x20 会按 0 处理，保证寄存器位置不漂移
func (df *DecodedFrame) RawPayload() ([]byte, bool) {
	if len(df.INFO) == 0 {
		return nil, false
	}
	b := asciiHexDecodePairs(df.INFO)
	if len(b) == 0 {
		return nil, false
	}
	return b, true
}

// —— 新增：把任意 ASCII hex（含空格占位）按“两个字符=一个字节”解码 —— //

func hexNibble(b byte) byte {
	switch {
	case b >= '0' && b <= '9':
		return b - '0'
	case b >= 'a' && b <= 'f':
		return b - 'a' + 10
	case b >= 'A' && b <= 'F':
		return b - 'A' + 10
	default:
		// 参照 原版 行为：遇到非十六进制字符（如 0x20 空格）按 0 处理
		return 0
	}
}

// asciiHexDecodePairs: 按“两字符一字节”转换；对非 hex 字符用 0 填充，保证位置不偏移
func asciiHexDecodePairs(ascii []byte) []byte {
	if len(ascii) < 2 {
		return nil
	}
	n := len(ascii) / 2
	out := make([]byte, 0, n)
	for i := 0; i+1 < len(ascii); i += 2 {
		hi := hexNibble(ascii[i])
		lo := hexNibble(ascii[i+1])
		out = append(out, (hi<<4)|lo)
	}
	return out
}
