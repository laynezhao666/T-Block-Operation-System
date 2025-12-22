package utils

import (
	"crypto/md5"
	"encoding/hex"
)

// Md5 md5加密
func Md5(str string) string {
	hash1 := md5.Sum([]byte(str))
	md5 := hex.EncodeToString(hash1[:])
	return md5
}
