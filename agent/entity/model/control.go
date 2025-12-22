package model

import "agent/entity/definition"

const (
	ReplyCodeInit       = -1000
	ChannelNotFind      = -1001
	ControlTimeOut      = -1002
	DeviceNotFind       = -1003
	ControlPointNotFind = -1004
)

// PointControlInfo 测点控制信息
type PointControlInfo struct {
	DeviceId  definition.DeviceGidType
	DeviceGid definition.DeviceGidType
	PointNo   definition.DataPointIDType
	PointGid  definition.DataPointIDType
	Value     string
}
