package data

import (
	"agent/entity/definition"
	"agent/logic/collector/rtdb/model"
)

// DataUnit 数据分发单元
type DataUnit struct {
	DeviceGid definition.DeviceGidType `json:"id"`
	// 测点数据
	Points model.DataPoints `json:"points"`
}

// KafkaPoint 写入 kafka 的测点
type KafkaPoint struct {
	ID        string `json:"i"`
	Value     string `json:"v"`
	Quality   string `json:"q"`
	Timestamp string `json:"t"`
}
