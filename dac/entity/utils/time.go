// Package utils 提供门禁系统通用工具函数。
package utils

import (
	"time"

	"dac/entity/consts"

	"dac/entity/utils/ttime"
)

// historyBeginTimestamp 历史数据起始时间戳
var (
	historyBeginTimestamp int64
)

// init 初始化历史数据起始时间戳
func init() {
	t, err := ttime.ParseLocal("2024-01-01 00:00:00")
	if err != nil {
		panic(err)
	}
	historyBeginTimestamp = t.Unix()
}

// GetHistoryBeginTimestamp 获取历史数据起始时间戳（秒）
func GetHistoryBeginTimestamp() int64 {
	return historyBeginTimestamp
}

// getWaitTime 从扩展配置中获取等待时间，若未配置则使用默认值
func getWaitTime(extend map[string]interface{},
	key string, defaultWaitTime time.Duration,
) time.Duration {
	var waitTime = defaultWaitTime
	for {
		v, ok := extend[key]
		if !ok {
			break
		}
		switch vv := v.(type) {
		case int:
			waitTime = time.Duration(vv) * time.Millisecond
		case float64:
			waitTime = time.Duration(vv) * time.Millisecond
		case float32:
			waitTime = time.Duration(vv) * time.Millisecond
		default:
			break
		}
		break
	}
	if waitTime <= 0 {
		waitTime = defaultWaitTime
	}
	return waitTime
}

// GetEventFetchWaitTime 获取事件拉取间隔时间
func GetEventFetchWaitTime(
	extend map[string]interface{},
) time.Duration {
	return getWaitTime(extend,
		consts.KeyFetchEventInterval,
		consts.DefaultEventFetchWaitTime)
}

// GetEventFetchLoopWaitTime 获取事件循环拉取间隔时间
func GetEventFetchLoopWaitTime(
	extend map[string]interface{},
) time.Duration {
	return getWaitTime(extend,
		consts.KeyFetchLoopEventInterval,
		consts.DefaultEventLoopWaitTime)
}

// GetAlarmFetchWaitTime 获取告警拉取间隔时间
func GetAlarmFetchWaitTime(
	extend map[string]interface{},
) time.Duration {
	return getWaitTime(extend,
		consts.KeyFetchAlarmInterval,
		consts.DefaultAlarmFetchWaitTime)
}

// GetAlarmFetchLoopWaitTime 获取告警循环拉取间隔时间
func GetAlarmFetchLoopWaitTime(
	extend map[string]interface{},
) time.Duration {
	return getWaitTime(extend,
		consts.KeyFetchLoopAlarmInterval,
		consts.DefaultAlarmLoopWaitTime)
}
