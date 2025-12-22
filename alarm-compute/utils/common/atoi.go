package common

import (
	"strconv"
)

// Str2Int Str2Int
func Str2Int(s string) int {
	v, _ := strconv.ParseInt(s, 10, 0)
	return int(v)
}

// Str2Uint Str2Uint
func Str2Uint(s string) uint {
	v, _ := strconv.ParseUint(s, 10, 0)
	return uint(v)
}

// Str2Int64 Str2Int64
func Str2Int64(s string) int64 {
	v, _ := strconv.ParseInt(s, 10, 64)
	return v
}

// Str2Uint64 Str2Uint64
func Str2Uint64(s string) uint64 {
	v, _ := strconv.ParseUint(s, 10, 64)
	return v
}

// StrToInt32 StrToInt32
func StrToInt32(s string) int32 {
	v, _ := strconv.ParseInt(s, 10, 0)
	return int32(v)
}

// StrToUint32 StrToUint32
func StrToUint32(s string) uint32 {
	v, _ := strconv.ParseUint(s, 10, 0)
	return uint32(v)
}
