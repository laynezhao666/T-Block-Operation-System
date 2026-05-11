package dianzong

import (
	"agent/entity/consts"
	"agent/entity/definition"
	model3 "agent/entity/model"
	"agent/logic/collector/device/model"
	"agent/utils"
	"context"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"

	"trpc.group/trpc-go/trpc-go/log"
)

// 通道类型常量
const (
	ChannelTypeSerial = "serial"
	ChannelTypeSocket = "socket"
)

// 扩展参数名称
const (
	ExtendParamEOI       = "EOI"        // 按帧头帧尾接收（不按长度）
	ExtendParamIgnoreRTN = "IGNORE_RTN" // 忽略RTN错误码检查
)

// Device 实现 IDevice
type Device struct {
	gid  definition.DeviceGidType
	name string

	channel IChannel // 通道接口（串口或TCP）
	addr    byte     // 设备地址
	ver     byte     // 协议版本

	// 配置选项
	receiveByEOI bool // true=按SOI/EOI接收，false=按长度接收
	ignoreRTN    bool // 是否忽略RTN错误码
	timeoutMs    int  // 超时时间
}

// Open 打开通道，等待发送指令
func (d *Device) Open(chanInfo model.ChannelInfo, packets model.ListCollectPackets) consts.Quality {
	// 解析扩展参数
	d.parseExtendParams(chanInfo)

	// 超时设置
	d.timeoutMs = utils.FirstNonZero(chanInfo.TimeoutMs, 3000)

	// 解析设备地址
	if len(chanInfo.Address) > 0 {
		b, err := parseProtoByte(chanInfo.Address)
		if err != nil {
			log.Errorf("chanInfo.Address parse err:%v", err)
			return consts.QualityConfigError
		}
		d.addr = b
	} else {
		log.Errorf("chanInfo.Address empty")
		return consts.QualityConfigError
	}

	// 解析协议版本
	if len(chanInfo.ProtocolVer) > 0 {
		b, err := parseProtoByte(chanInfo.ProtocolVer)
		if err != nil {
			log.Errorf("chanInfo.ProtocolVer parse err:%v", err)
			return consts.QualityConfigError
		}
		d.ver = b
	} else {
		log.Errorf("chanInfo.ProtocolVer empty")
		return consts.QualityConfigError
	}

	// 根据通道类型创建通道
	chType := strings.ToLower(chanInfo.ChType)
	if chType == "" {
		// 自动判断：如果Name包含":"则认为是TCP，否则是串口
		if strings.Contains(chanInfo.Name, ":") {
			chType = ChannelTypeSocket
		} else {
			chType = ChannelTypeSerial
		}
	}

	var err error
	switch chType {
	case ChannelTypeSocket:
		err = d.openTCPChannel(chanInfo)
	default:
		err = d.openSerialChannel(chanInfo)
	}

	if err != nil {
		log.Errorf("dianzong open channel err:%v", err)
		return consts.QualityCannotOpen
	}

	log.Infof("dianzong device opened: gid=%v type=%s addr=0x%02X ver=0x%02X receiveByEOI=%v",
		d.gid, chType, d.addr, d.ver, d.receiveByEOI)

	return consts.QualityOk
}

// openSerialChannel 打开串口通道
func (d *Device) openSerialChannel(chanInfo model.ChannelInfo) error {
	baudRate, parity, dataBits, stopBits, err := model.ParseRTUParam(chanInfo.Params)
	if err != nil {
		log.Errorf("param parse err:%v, use default 9600:N:8:1!!", err)
		baudRate = 9600
		parity = "N"
		dataBits = 8
		stopBits = 1
	}

	opts := SerialOptions{
		Port:       chanInfo.Name,
		Baud:       baudRate,
		DataBits:   dataBits,
		StopBits:   stopBits,
		Parity:     model.NormalizeParity(parity),
		ReadTO:     time.Duration(d.timeoutMs) * time.Millisecond,
		WriteTO:    2 * time.Second,
		HardwareFC: false,
		SoftwareFC: false,
	}

	sp := NewSerialChannel(opts)
	if err := sp.Open(); err != nil {
		return fmt.Errorf("serial open failed: %w", err)
	}
	d.channel = sp
	return nil
}

// openTCPChannel 打开TCP通道
func (d *Device) openTCPChannel(chanInfo model.ChannelInfo) error {
	opts := TCPOptions{
		Address: chanInfo.Name,
		Timeout: time.Duration(d.timeoutMs) * time.Millisecond,
	}

	tc := NewTCPChannel(opts)
	if err := tc.Open(); err != nil {
		return fmt.Errorf("tcp open failed: %w", err)
	}
	d.channel = tc
	return nil
}

// parseExtendParams 解析扩展参数
func (d *Device) parseExtendParams(chanInfo model.ChannelInfo) {
	extend := chanInfo.DriverExtend

	// 检查是否使用EOI方式接收（按帧头帧尾）
	// 串口默认使用EOI方式，TCP默认使用长度方式
	chType := strings.ToLower(chanInfo.ChType)
	if chType == ChannelTypeSocket {
		d.receiveByEOI = false // TCP默认按长度接收
	} else {
		d.receiveByEOI = true // 串口默认按帧标识接收
	}

	// 从扩展参数覆盖默认值
	if strings.Contains(extend, ExtendParamEOI) {
		d.receiveByEOI = true
	}
	if strings.Contains(extend, ExtendParamIgnoreRTN) {
		d.ignoreRTN = true
	}

	// 也检查ExtendKV
	if chanInfo.ExtendKV != nil {
		if _, has := chanInfo.ExtendKV["eoi"]; has {
			d.receiveByEOI = true
		}
		if _, has := chanInfo.ExtendKV["ignore_rtn"]; has {
			d.ignoreRTN = true
		}
	}
}

// Close 关闭通道
func (d *Device) Close() consts.Quality {
	if d.channel != nil {
		_ = d.channel.Close()
		d.channel = nil
	}
	return consts.QualityOk
}

// Request 发送采集指令，并根据响应将解析后的数据填充到测点值
func (d *Device) Request(ctx context.Context, packet *model.CollectProtocolPacket) (consts.Quality, model3.MessageStatistics) {
	var stat model3.MessageStatistics

	if d.channel == nil {
		return consts.QualityCommDisconnected, stat
	}

	// 检查通道是否打开，如果未打开则尝试重新打开
	if !d.channel.IsOpen() {
		if err := d.channel.ReOpen(); err != nil {
			log.Errorf("dianzong reopen channel err:%v", err)
			return consts.QualityCommDisconnected, stat
		}
	}

	log.Infof("dianzong build input: gid=%v VER=0x%02X ADR=0x%02X CMD=%s", d.gid,
		d.ver, d.addr, normalizeCmdHex(packet.Command))

	req, parsePlan, err := BuildRequestFromPacket(d.ver, d.addr, packet)
	if err != nil {
		log.Errorf("BuildRequestFromPacket err:%v", err)
		return consts.QualityConfigError, stat
	}

	log.Infof("dianzong send: gid=%v %s", d.gid, hexSpace(req))

	// 执行通信交换
	resp, err := d.exchange(ctx, req)
	respHex := toHex(resp)
	stat.RecvPackets = respHex
	if err != nil {
		log.Errorf("exchange gid=%v err:%v", d.gid, err)
		stat.ErrLog = err
		return consts.QualityCmdSendError, stat
	}

	log.Infof("dianzong recv: gid=%v %s\n", d.gid, respHex)

	// 解析响应
	decoded, err := ParseResponse(resp)
	if err != nil {
		log.Errorf("ParseResponse err:%v", err)
		// 解析失败视为有"异常/脏数据"，立刻清空以免污染下一次
		d.drainAndLog("parse-error")
		return consts.QualityCmdRespError, stat
	}

	// 检查RTN（如果未配置忽略）
	if !d.ignoreRTN && !decoded.OK {
		log.Errorf("RTN error: 0x%02X", decoded.RTN)
		return consts.QualityConfigError, stat
	}

	// 将数据映射到测点（使用 ValueParser）
	if parsePlan != nil {
		if err := parsePlan(decoded, packet); err != nil {
			// 某些寄存器越界/暂不支持是预期内的，降级为Debug级别避免大量日志
			log.Debugf("parsePlan err:%v", err)
			return consts.QualityCmdRespError, stat
		}
	}

	return consts.QualityOk, stat
}

func toHex(b []byte) string {
	const hexdigits = "0123456789ABCDEF"
	out := make([]byte, 0, len(b)*3)
	for i, v := range b {
		if i > 0 {
			out = append(out, ' ')
		}
		out = append(out, hexdigits[v>>4], hexdigits[v&0x0F])
	}
	return string(out)
}

// RequestPing 发送采集指令，最小化指令发送包
func (d *Device) RequestPing(ctx context.Context, packet model.CollectProtocolPacket) consts.Quality {
	q, _ := d.Request(ctx, &packet)
	return q
}

// Control 发送控制指令
func (d *Device) Control(packet *model.ControlProtocolPacket, val string) consts.Quality {
	return consts.QualityOk
}

// —— 内部辅助 —— //

func (d *Device) drainAndLog(why string) {
	if d.channel == nil {
		return
	}
	drained := d.channel.Drain(40*time.Millisecond, 4096)
	if len(drained) > 0 {
		log.Infof("channel drained after %s: %dB [%s]", why, len(drained), hexSpace(drained))
	}
}

func (d *Device) exchange(ctx context.Context, payload []byte) ([]byte, error) {
	// 发送前：尽量清掉上一次残留
	d.drainAndLog("pre-write")

	// 发送
	if err := d.channel.Write(payload); err != nil {
		// 发送失败也清一把，避免部分发送导致对端回应半截
		d.drainAndLog("write-error")
		return nil, err
	}

	// 根据配置选择接收方式
	var buf []byte
	var err error
	if d.receiveByEOI {
		// 按SOI/EOI帧标识接收
		buf, err = d.channel.ReadFrame(ctx)
	} else {
		// 按长度接收（先读头部，解析长度后读剩余）
		buf, err = d.channel.ReadByLength(ctx, FirstRecvLen, ParseRemainLenFromHead)
	}

	if err != nil {
		// 读失败时清掉可能残留的半帧/噪声
		d.drainAndLog("read-error")
		return nil, err
	}
	return buf, nil
}

// 解析一个"配置里的字节"，优先按十六进制，失败再回退十进制。
// 支持: "10" / "0x10" / "0A" / "a" / "01" / "255" 等。
func parseProtoByte(s string) (byte, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("empty byte string")
	}
	// —— 优先尝试十六进制 —— //
	if b, ok := tryParseHexOneByte(s); ok {
		return b, nil
	}
	// —— 回退十进制 —— //
	v, err := strconv.Atoi(s)
	if err != nil || v < 0 || v > 255 {
		return 0, fmt.Errorf("invalid byte value: %q", s)
	}
	return byte(v), nil
}

// 只解析"恰好一字节"的十六进制；对奇数字符补 0 前缀。
// 返回 (值, 是否按十六进制成功)
func tryParseHexOneByte(s string) (byte, bool) {
	c := strings.ToLower(strings.TrimSpace(s))
	c = strings.ReplaceAll(c, "0x", "")
	c = strings.ReplaceAll(c, " ", "")
	c = strings.ReplaceAll(c, "_", "")
	c = strings.ReplaceAll(c, "-", "")
	if c == "" {
		return 0, false
	}
	if len(c)%2 != 0 {
		c = "0" + c // "a" -> "0a"
	}
	if len(c) != 2 {
		return 0, false // 超过 1 字节则不当作十六进制解析
	}
	dst := make([]byte, 1)
	if _, err := hex.Decode(dst, []byte(c)); err != nil {
		return 0, false
	}
	return dst[0], true
}

// 把命令字符串规范化为可读十六进制：去空格/0x/_/-，再按 2 个一组输出（分组+空格）
func normalizeCmdHex(s string) string {
	c := strings.TrimSpace(s)
	c = strings.ReplaceAll(c, "0x", "")
	c = strings.ReplaceAll(c, "0X", "")
	c = strings.ReplaceAll(c, " ", "")
	c = strings.ReplaceAll(c, "_", "")
	c = strings.ReplaceAll(c, "-", "")
	c = strings.ToUpper(c)
	// 两位分组加空格
	var sb strings.Builder
	for i := 0; i < len(c); i += 2 {
		if i > 0 {
			sb.WriteByte(' ')
		}
		if i+2 <= len(c) {
			sb.WriteString(c[i : i+2])
		} else {
			sb.WriteByte(c[i]) // 奇数长度也兜底打印
		}
	}
	return sb.String()
}

// 把原始字节切片打印为 "AA BB CC" 形式
func hexSpace(b []byte) string {
	if len(b) == 0 {
		return ""
	}
	var sb strings.Builder
	for i, v := range b {
		if i > 0 {
			sb.WriteByte(' ')
		}
		sb.WriteString(fmt.Sprintf("%02X", v))
	}
	return sb.String()
}
