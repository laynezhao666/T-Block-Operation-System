// Package worker worker
package worker

import (
	"agent/entity/definition"
	"agent/entity/model"
	"agent/utils/message"
)

// DeviceMessage 设备变更消息
type DeviceMessage struct {
	Info model.DeviceInfo
	message.NoticeMessage
}

// NewDeviceMessage 组合设备配置信息，部分字段赋默认值
func NewDeviceMessage(method message.MethodType, info model.DeviceInfo) *DeviceMessage {
	return &DeviceMessage{
		Info:          info,
		NoticeMessage: *message.NewNoticeMessage(message.TopicDevice, method),
	}
}

// NewDeviceDeleteInChannelMessage 设备删除消息
func NewDeviceDeleteInChannelMessage(deviceGid definition.DeviceGidType, channel string) *DeviceMessage {
	return NewDeviceMessage(message.MethodDelete, model.DeviceInfo{Gid: deviceGid, ChannelID: channel})
}

// PointControlMessage 测点控制消息
type PointControlMessage struct {
	CtlInfo   model.PointControlInfo
	ReplyCode int
	message.NoticeMessage
}

// NewPointControlMessage 测点控制信息
func NewPointControlMessage(method message.MethodType, info model.PointControlInfo) *PointControlMessage {
	return &PointControlMessage{
		CtlInfo:       info,
		ReplyCode:     model.ReplyCodeInit,
		NoticeMessage: *message.NewNoticeMessage(message.TopicPointControl, method),
	}
}
