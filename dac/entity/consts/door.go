// Package consts 定义门禁系统的全局常量。
package consts

// 门参数默认值
const (
	DefaultDoorKeepOpenTime   = 5   // 默认门保持开启时间（秒）
	DefaultDoorOpenTimeout    = 15  // 默认开门超时时间（秒）
	DefaultDoorLockCount      = 5   // 默认锁定次数
	DefaultDoorLockTime       = 300 // 默认锁定时间（秒）
	DefaultDoorVerifyInterval = 60  // 默认验证间隔（秒）
	DefaultDoorOpenMode       = 0   // 默认开门模式
	DefaultDoorFireSignalMode = 0   // 默认消防信号模式
)
