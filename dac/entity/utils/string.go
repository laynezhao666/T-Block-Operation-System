// Package utils 提供门禁系统通用工具函数。
package utils

import (
	"fmt"
	"strconv"
	"time"
)

// ToHex 将字节切片转换为十六进制字符串，使用指定分隔符连接
func ToHex(buf []byte, rep string) string {
	var res string
	if buf == nil || len(buf) <= 0 {
		return res
	}
	res += fmt.Sprintf("%02X", buf[0])
	for i := 1; i < len(buf); i++ {
		res += rep
		res += fmt.Sprintf("%02X", buf[i])
	}
	return res
}

// StringToUint8 将字符串转换为uint8类型
func StringToUint8(str string) (uint8, error) {
	i, err := strconv.Atoi(str)
	if err != nil {
		return 0, err
	}
	return uint8(i), nil
}

// StringToUnixTime 将时间字符串解析为Unix时间戳（秒）
func StringToUnixTime(str string) (int64, error) {
	t, err := time.ParseInLocation(
		"2006-01-02 15:04:05", str, time.Local)
	if err != nil {
		return 0, err
	}
	return t.Unix(), nil
}
