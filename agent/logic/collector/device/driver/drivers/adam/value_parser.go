package adam

import (
	"fmt"
	"strings"
)

// ValueParser 解析器
//   - Addr:     返回数据的字节偏移；例如传 0 表示用第 0 字节
//   - DataType: 支持 "BOOL0"~"BOOL7"（取该字节的第1~第8位）；
type ValueParser struct {
	Addr     uint32
	Extend   string
	DataType string
}

// ReadFrom 从 payload 抽取值
func (vp *ValueParser) ReadFrom(payload []byte) (any, error) {
	if vp == nil {
		return nil, fmt.Errorf("nil ValueParser")
	}
	if int(vp.Addr) >= len(payload) {
		return nil, fmt.Errorf("addr out of range: %d >= %d", vp.Addr, len(payload))
	}
	b := payload[vp.Addr]

	switch strings.ToUpper(strings.TrimSpace(vp.DataType)) {
	case "BOOL0", "BOOL1", "BOOL2", "BOOL3", "BOOL4", "BOOL5", "BOOL6", "BOOL7":
		bit := uint(strings.TrimPrefix(strings.ToUpper(vp.DataType), "BOOL")[0] - '0') // 0..7
		return int((b >> bit) & 0x01), nil
	case "BYTE", "U8", "UINT8":
		return uint8(b), nil
	default:
		// 未配置 DataType 时，默认取 BYTE
		if strings.TrimSpace(vp.DataType) == "" {
			return uint8(b), nil
		}
		return nil, fmt.Errorf("unsupported DataType: %s", vp.DataType)
	}
}
