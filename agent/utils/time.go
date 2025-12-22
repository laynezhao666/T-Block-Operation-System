package utils

import (
	"time"
)

var (
	loc *time.Location
)

func init() {
	loc = time.FixedZone("CST", 8*3600) // 避免环境中缺少时区信息，手动设置中国时区

}

// GetNowLocalTime 获取当前时间
func GetNowLocalTime() time.Time {
	return time.Now()
}

// GetNowBeijingTime 获取当前北京时间
func GetNowBeijingTime() time.Time {
	return time.Now().In(loc)
}

// GetNowUTCTime 获取当前UTC时间
func GetNowUTCTime() time.Time {
	return time.Now().UTC()
}

// GetNowUTCTimeStamp 获取当前UTC时间戳
func GetNowUTCTimeStamp() int64 {
	return time.Now().Unix()
}

// ConvertIfMilliToSeconds 如果是毫秒时间戳则转换为秒时间戳
func ConvertIfMilliToSeconds(ts int64) int64 {
	if ts > 1e11 {
		return ts / 1000
	}
	return ts
}
