// Package ttime 提供统一的时间获取工具函数，
// 封装UTC和本地时间的获取方法。
package ttime

import (
	"time"
)

// GetNowUTC 获取当前UTC时间
func GetNowUTC() time.Time {
	return time.Now().UTC()
}

// GetNowLocal 获取当前本地时间
func GetNowLocal() time.Time {
	return time.Now()
}
