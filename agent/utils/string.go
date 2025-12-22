package utils

import (
	"strconv"
	"strings"
)

// LoadHex 将十六进制字符串转换为字节数组
func LoadHex(buf []byte, bufLen *int, s string) bool {
	hexStr := strings.TrimSpace(s)
	lenHex := len(hexStr)

	// 检查输入参数
	if buf == nil || lenHex%2 != 0 {
		return false
	}

	j := 0
	for i := 0; i < lenHex; i += 2 {
		// 解析两个字符为一个字节
		value, err := strconv.ParseUint(hexStr[i:i+2], 16, 8)
		if err != nil {
			return false
		}
		buf[j] = byte(value)
		j++
	}

	// 设置 bufLen
	if bufLen != nil {
		*bufLen = j
	}

	return true
}
