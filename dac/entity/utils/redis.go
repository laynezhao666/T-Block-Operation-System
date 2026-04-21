// Package utils 提供门禁系统通用工具函数。
package utils

import (
	"fmt"

	chdConsts "dac/logic/collect/driver/chd806d4/consts"
	"dac/logic/collect/driver/xbrother/consts"
)

// GenerateRedisKeyDoorOpenTimeout 生成XBrother门开超时告警的Redis Key
func GenerateRedisKeyDoorOpenTimeout(channelID string) string {
	return fmt.Sprintf("%s:%s:%s",
		consts.RedisKeyXBrotherPredix, "current_alarm", channelID)
}

// GenerateRedisKeyDoorStatus 生成XBrother门状态的Redis Key
func GenerateRedisKeyDoorStatus(channelID string) string {
	return fmt.Sprintf("%s:%s:%s",
		consts.RedisKeyXBrotherPredix, "door_status", channelID)
}

// ============ CHD 协议专用 Redis Key ============

// GenerateRedisKeyCHDDoorStatus CHD 门状态 Redis Key
func GenerateRedisKeyCHDDoorStatus(channelID string) string {
	return fmt.Sprintf("%s:%s:%s",
		chdConsts.RedisKeyCHDPrefix, "door_status", channelID)
}

// GenerateRedisKeyCHDCurrentAlarm CHD 当前告警 Redis Key
func GenerateRedisKeyCHDCurrentAlarm(channelID string) string {
	return fmt.Sprintf("%s:%s:%s",
		chdConsts.RedisKeyCHDPrefix, "current_alarm", channelID)
}
