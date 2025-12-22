// Package stringutil provides various string tools
package stringutil

import (
	"fmt"
	"reflect"
	"strings"
	"unicode/utf8"

	"github.com/asaskevich/govalidator"
)

// Diff Creates an slice of slice values not included in the other given slice.
func Diff(base, exclude []string) (result []string) {
	excludeMap := make(map[string]bool)
	for _, s := range exclude {
		excludeMap[s] = true
	}
	for _, s := range base {
		if !excludeMap[s] {
			result = append(result, s)
		}
	}
	return result
}

// Unique 切片去重
func Unique(ss []string) (result []string) {
	smap := make(map[string]bool)
	for _, s := range ss {
		smap[s] = true
	}
	for s := range smap {
		result = append(result, s)
	}
	return result
}

// CamelCaseToUnderscore 驼峰命名转下划线命名
func CamelCaseToUnderscore(str string) string {
	return govalidator.CamelCaseToUnderscore(str)
}

// UnderscoreToCamelCase 下划线命名转驼峰命名
func UnderscoreToCamelCase(str string) string {
	return govalidator.UnderscoreToCamelCase(str)
}

// FindString 切片查找字符串
func FindString(array []string, str string) int {
	for index, s := range array {
		if str == s {
			return index
		}
	}
	return -1
}

// StringIn 字符串切片判断元素存在
func StringIn(str string, array []string) bool {
	return FindString(array, str) > -1
}

// Reverse 字符串逆序
func Reverse(s string) string {
	size := len(s)
	buf := make([]byte, size)
	for start := 0; start < size; {
		r, n := utf8.DecodeRuneInString(s[start:])
		start += n
		utf8.EncodeRune(buf[size-start:], r)
	}
	return string(buf)
}

// QueryString convert map to string
func QueryString(data map[string]interface{}) string {
	var sb strings.Builder
	idx := 1
	for k, v := range data {
		sb.WriteString(k)
		sb.WriteString("=")
		// bool 转字符串
		switch value := v.(type) {
		case bool:
			if value {
				sb.WriteString("1")
			} else {
				sb.WriteString("0")
			}
		default:
			sb.WriteString(fmt.Sprintf("%v", v))
		}
		// 非最后才加
		if idx < len(data) {
			sb.WriteString("&")
		}
		idx++
	}
	return sb.String()
}

// Capitalize  字符串首字母转大写
func Capitalize(s string) string {
	var upperStr string
	vv := []rune(s)
	for i := 0; i < len(vv); i++ {
		if i == 0 {
			if vv[i] >= 97 && vv[i] <= 122 { //小写字母的Unicode编码
				vv[i] -= 32 // string的码表相差32位
				upperStr += string(vv[i])
			} else { //首字母非小写字母，直接返回原字符串
				return s
			}
		} else {
			upperStr += string(vv[i])
		}
	}
	return upperStr
}

// SubString 截取字符串，支持中文截取
func SubString(s string, begin, length int) (substr string) {
	// 将字符串的转换成[]rune
	rs := []rune(s)
	lth := len(rs)

	// 简单的越界判断
	if begin < 0 {
		begin = 0
	}
	if begin >= lth {
		begin = lth
	}
	end := begin + length
	if end > lth {
		end = lth
	}

	// 返回子串
	return string(rs[begin:end])
}

// Join 用字符串连接数组元素
func Join(elems interface{}, sep string) string {
	if elems == nil {
		return ""
	}

	if reflect.TypeOf(elems).Kind() != reflect.Array && reflect.TypeOf(elems).Kind() != reflect.Slice {
		// TODO: 考虑打印日志或返回error
		return ""
	}

	s := reflect.ValueOf(elems)

	var strElems []string
	strElems = make([]string, s.Len())
	for i := 0; i < s.Len(); i++ {
		strElems[i] = fmt.Sprintf("%v", s.Index(i).Interface())
	}

	return strings.Join(strElems, sep)
}

// SplitInt 使用一个字符串分割另一个字符串，返回 int 数组
func SplitInt(s, sep string) []int {
	strArr := strings.Split(s, sep)

	ret := make([]int, len(strArr))

	for k, v := range strArr {
		ret[k] = StringToInt(v)
	}

	return ret
}

// SplitInt64 使用一个字符串分割另一个字符串，返回 int64 数组
func SplitInt64(s, sep string) []int64 {
	strArr := strings.Split(s, sep)

	ret := make([]int64, len(strArr))

	for k, v := range strArr {
		ret[k] = StringToInt64(v)
	}

	return ret
}

// SplitUint16 使用一个字符串分割另一个字符串，返回 uint16 数组
func SplitUint16(s, sep string) []uint16 {
	strArr := strings.Split(s, sep)

	ret := make([]uint16, len(strArr))

	for k, v := range strArr {
		ret[k] = StringToUint16(v)
	}

	return ret
}

// SplitUint32 使用一个字符串分割另一个字符串，返回 uint32 数组
func SplitUint32(s, sep string) []uint32 {
	strArr := strings.Split(s, sep)

	ret := make([]uint32, len(strArr))

	for k, v := range strArr {
		ret[k] = StringToUint32(v)
	}

	return ret
}

// SplitUint64 使用一个字符串分割另一个字符串，返回 uint64 数组
func SplitUint64(s, sep string) []uint64 {
	strArr := strings.Split(s, sep)

	ret := make([]uint64, len(strArr))

	for k, v := range strArr {
		ret[k] = StringToUint64(v)
	}

	return ret
}
