package ttime

import (
	"time"
)

// TruncateMinute 将当前时间戳向下取整到所在分钟的第 0s 对应时间戳
func TruncateMinute(timestampSecond int64) int64 {
	return time.Unix(timestampSecond, 0).Truncate(time.Minute).Unix()
}
