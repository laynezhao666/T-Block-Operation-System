package common

import (
	"strconv"
)

// Int2Str Int2Str
func Int2Str(v int) string {
	return strconv.FormatInt(int64(v), 10)
}

// Uint2Str Uint2Str
func Uint2Str(v uint) string {
	return strconv.FormatUint(uint64(v), 10)
}

// Int642Str Int642Str
func Int642Str(v int64) string {
	return strconv.FormatInt(v, 10)
}

// Uint642Str Uint642Str
func Uint642Str(v uint64) string {
	return strconv.FormatUint(v, 10)
}

// Int32ToStr Int32ToStr
func Int32ToStr(v int32) string {
	return strconv.FormatInt(int64(v), 10)
}

// Uint32ToStr Uint32ToStr
func Uint32ToStr(v uint32) string {
	return strconv.FormatUint(uint64(v), 10)
}
