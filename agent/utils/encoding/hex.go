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
