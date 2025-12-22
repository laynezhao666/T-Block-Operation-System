package encoding

import (
	"crypto/md5"
	"encoding/hex"
	"os"
)

// MD5 returns the MD5 checksum of data.
func MD5(data []byte) string {
	b := md5.Sum(data)
	return hex.EncodeToString(b[:])
}

// MD5String returns the MD5 checksum of s.
func MD5String(s string) string {
	return MD5([]byte(s))
}

// MD5File returns the MD5 checksum of a file.
func MD5File(filename string) (string, error) {
	b, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}
	return MD5(b), nil
}
