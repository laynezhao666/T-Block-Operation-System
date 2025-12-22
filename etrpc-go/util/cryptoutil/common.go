// Package cryptoutil provides ...
package cryptoutil

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"hash"
)

// Md5 md5算法
func Md5(data string) string {
	md := md5.Sum([]byte(data))
	return hex.EncodeToString(md[:])
}

// Sha1 sha1算法
func Sha1(data string) string {
	sha := sha1.Sum([]byte(data))
	return hex.EncodeToString(sha[:])
}

// Sha256 sha256算法
func Sha256(data string) string {
	sha256Hash := sha256.New()
	sha := sha256Hash.Sum([]byte(data))
	return hex.EncodeToString(sha[:])
}

// Sha512 sha512算法
func Sha512(data string) string {
	sha := sha512.New().Sum([]byte(data))
	return hex.EncodeToString(sha[:])
}

// Hmac hmac算法：algo 支持 sha1、sha256、sha512、md5 哈希算法
func Hmac(key, data, algo string) string {
	var f func() hash.Hash
	if algo == "sha1" {
		f = sha1.New
	} else if algo == "sha256" {
		f = sha256.New
	} else if algo == "sha512" {
		f = sha512.New
	} else {
		f = md5.New
	}

	h := hmac.New(f, []byte(key))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

// AesEncrypt AES加密，key的长度必须是16,24,32位
func AesEncrypt(key string, msg []byte) (ret []byte, err error) {
	bs := []byte(key)
	block, err := aes.NewCipher(bs)
	if err != nil {
		return nil, err
	}

	ret = make([]byte, len(msg))
	var iv = bs[:aes.BlockSize]
	encrypter := cipher.NewCFBEncrypter(block, iv)
	encrypter.XORKeyStream(ret, msg)

	return
}

// AesDecrypt AES 解密，key的长度必须是16,24,32位
func AesDecrypt(key string, src []byte) (ret []byte, err error) {
	defer func() {
		//错误处理
		if e := recover(); e != nil {
			err, _ = e.(error)
		}
	}()

	bs := []byte(key)

	block, err := aes.NewCipher(bs)
	if err != nil {
		return nil, err
	}

	ret = make([]byte, len(src))
	var iv = bs[:aes.BlockSize]
	decrypter := cipher.NewCFBDecrypter(block, iv)
	decrypter.XORKeyStream(ret, src)

	return
}

// EncodeBase64 base64编码
func EncodeBase64(data string) string {
	return base64.StdEncoding.EncodeToString([]byte(data))
}

// DecodeBase64 base64解码
func DecodeBase64(data string) (string, error) {
	b, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
