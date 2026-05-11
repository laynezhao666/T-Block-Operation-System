package dianzong

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
)

// ValueParser 解析器
type ValueParser struct {
	Addr      string
	Extend    string
	DataType  string //datatype.DataType
	ByteOrder string
}

// ReadByReg 从响应帧 df 中，按配置 (cid2, reg, datatype, byteorder) 抽取一个值。
//   - cid2: 来自配置的 cmd（例如 "2A44" => cid2 = 0x44）
//   - reg:  字节偏移（去掉 DATAFLAG 后的 payload 内的 0-based 下标），可写 "116" / "0x74"；
//     也支持位访问 "116:3"（取该字节 bit3，返回 0/1）。
//   - datatype: "uint8"/"int8"/"uint16"/"int16"/"uint32"/"int32"/"float32"/"bytes[n]" 等
//   - byteorder: "big" / "little"
func ReadByReg(df *DecodedFrame, cid2 byte, reg, datatype, byteorder string) (any, error) {
	// 1) 定位“去掉 DATAFLAG 后的 payload”
	flag, payload, err := df.payloadAfterFlag(cid2)
	_ = flag // 若上层需要可用 df.Parsed["DATAFLAG"]
	if err != nil {
		return nil, err
	}

	// 2) 解析 reg（支持 "N" / "0xNN" / "N:bit" / "0xNN:bit"）
	byteIdx, bitIdx, err := parseRegSpec(reg)
	if err != nil {
		return nil, err
	}
	if byteIdx < 0 || byteIdx >= len(payload) {
		return nil, fmt.Errorf("reg %s (byte %d) out of range, payload len=%d", reg, byteIdx, len(payload))
	}

	// 3) 大小端
	var order binary.ByteOrder = binary.BigEndian
	// float类型默认为little
	switch strings.ToLower(strings.TrimSpace(datatype)) {
	case "float", "double":
		order = binary.LittleEndian
	}

	if strings.EqualFold(byteorder, "little") {
		order = binary.LittleEndian
	}

	// 4) 位访问优先
	if bitIdx >= 0 {
		v := (payload[byteIdx] >> uint(bitIdx)) & 0x01
		return int(v), nil
	}

	// 5) 数据类型
	switch strings.ToLower(strings.TrimSpace(datatype)) {
	case "uint8":
		return uint8(payload[byteIdx]), nil
	case "int8":
		return int8(payload[byteIdx]), nil
	case "uint16":
		if byteIdx+1 >= len(payload) {
			return nil, errors.New("uint16 out of range")
		}
		return order.Uint16(payload[byteIdx : byteIdx+2]), nil
	case "int16":
		if byteIdx+1 >= len(payload) {
			return nil, errors.New("int16 out of range")
		}
		return int16(order.Uint16(payload[byteIdx : byteIdx+2])), nil
	case "uint32":
		if byteIdx+3 >= len(payload) {
			return nil, errors.New("uint32 out of range")
		}
		return order.Uint32(payload[byteIdx : byteIdx+4]), nil
	case "int32":
		if byteIdx+3 >= len(payload) {
			return nil, errors.New("int32 out of range")
		}
		return int32(order.Uint32(payload[byteIdx : byteIdx+4])), nil
	case "float":
		if byteIdx+3 >= len(payload) {
			return nil, errors.New("float32 out of range")
		}
		u := order.Uint32(payload[byteIdx : byteIdx+4])
		return math.Float32frombits(u), nil
	default:
		// 支持 bytes[n] 按原样返回
		if strings.HasPrefix(strings.ToLower(datatype), "bytes[") && strings.HasSuffix(datatype, "]") {
			nStr := strings.TrimSuffix(strings.TrimPrefix(datatype, "bytes["), "]")
			n, _ := strconv.Atoi(nStr)
			if n <= 0 || byteIdx+n > len(payload) {
				return nil, fmt.Errorf("bytes[%d] out of range", n)
			}
			out := make([]byte, n)
			copy(out, payload[byteIdx:byteIdx+n])
			return out, nil
		}
		return nil, fmt.Errorf("unsupported datatype: %s", datatype)
	}
}

// parseRegSpec: "116" / "0x74" / "116:3" / "0x74:3"
func parseRegSpec(s string) (byteIdx int, bitIdx int, err error) {
	bitIdx = -1
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, -1, errors.New("empty reg")
	}
	parts := strings.Split(s, ":")
	raw := parts[0]
	if len(parts) == 2 {
		bitIdx, err = strconv.Atoi(parts[1])
		if err != nil || bitIdx < 0 || bitIdx > 7 {
			return 0, -1, fmt.Errorf("bad bit index in reg: %s", s)
		}
	} else if len(parts) > 2 {
		return 0, -1, fmt.Errorf("bad reg spec: %s", s)
	}
	// 十进制或 0x.. 十六进制兼容
	if strings.HasPrefix(raw, "0x") || strings.HasPrefix(raw, "0X") {
		val, e := strconv.ParseUint(raw[2:], 16, 32)
		if e != nil {
			return 0, -1, e
		}
		return int(val), bitIdx, nil
	}
	// 纯十六进制（例如 "74" 也可被视为十进制；如需强制十六进制，可写 0x74）
	if _, e := strconv.Atoi(raw); e == nil {
		v, _ := strconv.Atoi(raw)
		return v, bitIdx, nil
	}
	// 兜底：尝试按 hex 解（如 "7A"）
	if _, e := hex.DecodeString(raw); e == nil && len(raw) <= 4 {
		val, e2 := strconv.ParseUint(raw, 16, 32)
		if e2 != nil {
			return 0, -1, e2
		}
		return int(val), bitIdx, nil
	}
	return 0, -1, fmt.Errorf("bad reg: %s", s)
}

//
//type ProtDef struct {
//	ByteOrder string `json:"byteorder"`
//	Cmd       string `json:"cmd"`
//	Datatype  string `json:"datatype"`
//	Ext       string `json:"ext"`
//	Reg       string `json:"reg"`
//}
//
//// 绑定到 ValueParser：从 Extend 里把 protdef 解出来（字段名差异也兼容几种常见大小写）
//func (vp *ValueParser) ParseProtDef() (*ProtDef, error) {
//	if vp == nil || strings.TrimSpace(vp.Extend) == "" {
//		return &pd, nil
//	}
//	// 直接把 Extend 当 JSON 解析
//	if err := json.Unmarshal([]byte(vp.Extend), &pd); err != nil {
//		// 兜底：把 "CMD=2A44;REG=74;DATATYPE=uint8;BYTEORDER=big" 这种 kv 也做个极简解析
//		items := strings.FieldsFunc(vp.Extend, func(r rune) bool { return r == ';' || r == ',' })
//		for _, it := range items {
//			kv := strings.SplitN(it, "=", 2)
//			if len(kv) != 2 {
//				continue
//			}
//			k, v := strings.ToLower(strings.TrimSpace(kv[0])), strings.TrimSpace(kv[1])
//			switch k {
//			case "cmd":
//				pd.Cmd = v
//			case "reg":
//				pd.Reg = v
//			case "datatype", "data_type":
//				pd.Datatype = v
//			case "byteorder", "endian":
//				pd.ByteOrder = v
//			}
//		}
//	}
//	return &pd, nil
//}
