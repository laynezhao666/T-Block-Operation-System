package processor

import (
	"agent/logic/collector/processor/iprocessor"
	"agent/logic/collector/processor/zero"
)

var (
	m iprocessor.Manager
)

// GetZeroManager 获取零值管理器
func GetZeroManager() iprocessor.Manager {
	return m
}

// Init 初始化
func Init() {
	// 对设备进行零值判定异常检测
	m = zero.NewManager()
}
