// Package encoding 提供常用编码工具函数。
package encoding

import (
	"crypto/md5"
	"encoding/hex"
	"os"
)

// MD5 计算字节切片的MD5哈希值，返回十六进制字符串
func MD5(data []byte) string {
	b := md5.Sum(data)
	return hex.EncodeToString(b[:])
}

// MD5String 计算字符串的MD5哈希值
func MD5String(s string) string {
	return MD5([]byte(s))
}

// MD5File 计算文件内容的MD5哈希值
func MD5File(filename string) (string, error) {
	b, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}
	return MD5(b), nil
}
