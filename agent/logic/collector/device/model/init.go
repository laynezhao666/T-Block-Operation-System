package model

import (
	"agent/entity/config"
)

var (
	maxAllowedPacketFailedCount = 2
)

// Init 初始化
func Init() {
	maxAllowedPacketFailedCount = config.LoadIntOrDefault(config.GetRB().Collector.Common.PacketFailedCount, 1)
}
