package encoding

import (
	"fmt"
	"strings"
)

// ParseBytesToHex 将字节切片转换为十六进制字符串（如 []byte{0, 248, 2, 158} -> "00 F8 02 9E"）
func ParseBytesToHex(data []byte) string {
	hexParts := make([]string, len(data))
	for i, b := range data {
		hexParts[i] = fmt.Sprintf("%02X", b)
	}
	return strings.Join(hexParts, " ")
}

// HexSp hex sp
func HexSp(b []byte) string {
	const h = "0123456789ABCDEF"
	if len(b) == 0 {
		return ""
	}
	out := make([]byte, 0, len(b)*3)
	for i, v := range b {
		if i > 0 {
			out = append(out, ' ')
		}
		out = append(out, h[v>>4], h[v&0x0F])
	}
	return string(out)
}
