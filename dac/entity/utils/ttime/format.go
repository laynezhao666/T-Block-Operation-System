// Package ttime 提供时间格式化和解析工具函数。
package ttime

import (
	"time"
)

// Layout 标准时间格式化模板
const (
	Layout = "2006-01-02 15:04:05"
)

// Format 将Unix秒级时间戳格式化为标准时间字符串
func Format(secondTimestamp int64) string {
	return time.Unix(secondTimestamp, 0).Format(Layout)
}

// Parse 将标准时间字符串解析为UTC时间
func Parse(t string) (time.Time, error) {
	return time.Parse(Layout, t)
}

// ParseLocal 将标准时间字符串解析为本地时区时间
func ParseLocal(t string) (time.Time, error) {
	return time.ParseInLocation(Layout, t, time.Local)
}
