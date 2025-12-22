// Package mysqlutil provides various mysql tools.
package mysqlutil

import "strings"

// MysqlRealEscapeString 数据库字符串字段内容过滤，防止sql注入
func MysqlRealEscapeString(value string) string {
	replace := map[string]string{"\\": "\\\\", "'": `\'`, "\\0": "\\\\0", "\n": "\\n", "\r": "\\r", `"`: `\"`,
		"\x1a": "\\Z"}

	for lodStr, newStr := range replace {
		value = strings.Replace(value, lodStr, newStr, -1)
	}

	return "'" + value + "'"
}

// MysqlRealEscapeStringV2 数据库字符串字段内容过滤，防止sql注入，不自动加单引号
func MysqlRealEscapeStringV2(value string) string {
	replace := map[string]string{"\\": "\\\\", "'": `\'`, "\\0": "\\\\0", "\n": "\\n", "\r": "\\r", `"`: `\"`,
		"\x1a": "\\Z"}

	for lodStr, newStr := range replace {
		value = strings.Replace(value, lodStr, newStr, -1)
	}

	return value
}
