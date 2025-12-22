package utils

import (
	"strings"
)

// ParseKvString 解析字符串，如 key1=value1,key2=value2,key3=value3, 返回 map[string]string
func ParseKvString(str string, multiSep string) map[string]string {
	result := make(map[string]string)

	items := strings.Split(str, multiSep)
	for i := range items {
		item := strings.Split(items[i], "=")
		if len(item) == 2 {
			result[item[0]] = item[1]
		}
	}
	return result
}
