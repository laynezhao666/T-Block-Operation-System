package model

import (
	"agent/entity/definition"
)

// CollectProtocolPacket 采集指令包
type CollectProtocolPacket struct {
	Command     string     // 采集指令
	Points      ListPoints // 采集测点
	failedCount int        // 失败次数
}

// GetFailedCount 获取失败次数
func (c *CollectProtocolPacket) GetFailedCount() int {
	if c == nil {
		return 0
	}
	return c.failedCount
}

// UpdateStat 更新统计信息
func (c *CollectProtocolPacket) UpdateStat(currentRequestSuccess bool) {
	if c == nil {
		return
	}
	if currentRequestSuccess {
		c.failedCount = 0
	} else {
		c.failedCount++
	}
}

// IsTimeout 判断是否超时
func (c *CollectProtocolPacket) IsTimeout() bool {
	if c == nil {
		return false
	}
	return c.failedCount > maxAllowedPacketFailedCount
}

// ListCollectPackets 采集指令包列表
type ListCollectPackets []*CollectProtocolPacket

func (l ListCollectPackets) getPointsNumber() int {
	n := 0
	for _, p := range l {
		n += len(p.Points)
	}
	return n
}

// GetPointIDs 获取测点ID列表
func (l ListCollectPackets) GetPointIDs() []definition.DataPointIDType {
	points := make([]definition.DataPointIDType, 0, l.getPointsNumber())
	for _, p := range l {
		for _, point := range p.Points {
			points = append(points, point.Attr.ID)
		}
	}
	return points
}

// GetPoints 获取采集测点列表
func (l ListCollectPackets) GetPoints() ListPoints {
	points := make(ListPoints, 0, l.getPointsNumber())
	for _, p := range l {
		points = append(points, p.Points...)
	}
	return points
}

// ControlProtocolPacket 控制指令包
type ControlProtocolPacket struct {
	Command string
	Point   *PointInfo
}
