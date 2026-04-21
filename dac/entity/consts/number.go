// Package consts 定义门禁系统的全局常量。
package consts

import (
	"time"
)

// 点位消息批量大小
const (
	PointNumberPerMessage = 50 // 每条消息包含的点位数量
)

// 默认超时时间
const (
	DefaultTimeoutMS = 10000 // 默认超时时间（毫秒）
)

// 默认账号密码
const (
	DefaultUserName = "system" // 默认用户名
	DefaultPassword = "666666" // 默认密码

	UnknownName = "未知" // 未知名称占位符
)

// 导出限制
const (
	MaxRecord = 100000 // 最大导出记录数
)

// 告警和事件采集间隔
const (
	DefaultAlarmFetchWaitTime = time.Second      // 告警采集等待时间
	DefaultAlarmLoopWaitTime  = time.Second * 45 // 告警循环等待时间
	DefaultEventFetchWaitTime = time.Second      // 事件采集等待时间
	DefaultEventLoopWaitTime  = time.Second * 45 // 事件循环等待时间
)
