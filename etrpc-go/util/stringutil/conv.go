// Package stringutil provides various string tools
package stringutil

import (
	"fmt"
	"strconv"
	"unicode/utf8"
)

// Uint64ToString uint64 -> string
func Uint64ToString(v uint64) string {
	return strconv.FormatUint(v, 10)
}

// Uint32ToString uint32 -> string
func Uint32ToString(v uint32) string {
	return fmt.Sprintf("%d", v)
}

// Uint16ToString uint16 -> string
func Uint16ToString(v uint16) string {
	return fmt.Sprintf("%d", v)
}

// StringToInt string -> int
func StringToInt(v string) int {
	ret, _ := strconv.Atoi(v)
	return ret
}

// StringToInt64 string -> int64
func StringToInt64(v string) int64 {
	ret, _ := strconv.ParseInt(v, 10, 64)
	return ret
}

// StringToUint16 string -> uint16
func StringToUint16(v string) uint16 {
	uintV, _ := strconv.ParseUint(v, 10, 16)
	return uint16(uintV)
}

// StringToUint32 string -> uint32
func StringToUint32(v string) uint32 {
	uintV, _ := strconv.ParseUint(v, 10, 32)
	return uint32(uintV)
}

// StringToUint64 string -> uint64
func StringToUint64(v string) uint64 {
	uintV, _ := strconv.ParseUint(v, 10, 64)
	return uintV
}

// StringUnicodeLen string自负长度
func StringUnicodeLen(v string) int {
	return utf8.RuneCountInString(v)
}

// StringToByte 免拷贝字符串转字节
func StringToByte(s string) (b []byte) {
	return []byte(s)
	//以下实现方式不安全
	//pbytes := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	//pstring := (*reflect.StringHeader)(unsafe.Pointer(&s))
	//pbytes.Data = pstring.Data
	//pbytes.Len = pstring.Len
	//pbytes.Cap = pstring.Len
	//return
}

// ByteToString 免拷贝字节转字符串
func ByteToString(b []byte) (s string) {
	return string(b)
	//以下实现方式不安全
	//pbytes := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	//pstring := (*reflect.StringHeader)(unsafe.Pointer(&s))
	//pstring.Data = pbytes.Data
	//pstring.Len = pbytes.Len
	//return
}

// Uint32ToChar int32 转换为
func Uint32ToChar(v uint32) string {
	return string(rune('A' - 1 + v))
}
