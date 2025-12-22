package iprocessor

import (
	"agent/entity/definition"
)

// Manager 管理接口
type Manager interface {
	// GetProcessor 获取处理器
	GetProcessor(deviceGiD definition.DeviceGidType, extends map[string]interface{}) Processor
}
