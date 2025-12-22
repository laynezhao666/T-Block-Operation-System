package iprocessor

import (
	"agent/entity/definition"
	"agent/logic/collector/rtdb/model"
)

// Processor 处理器接口
type Processor interface {
	// Do 处理
	Do(pointCount int, pointID definition.DataPointIDType, point *model.RTValue)
}
