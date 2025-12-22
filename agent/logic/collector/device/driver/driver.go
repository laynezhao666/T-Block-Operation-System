// Package driver 驱动层
package driver

import (
	"context"
	model2 "agent/entity/consts"
	"agent/entity/definition"
	model3 "agent/entity/model"
	"agent/logic/collector/device/model"
)

// IDriver 驱动对象
type IDriver interface {
	// Init 初始化驱动
	Init() model2.Quality
	// UnInit 反初始化驱动
	UnInit() model2.Quality
	// CreateDevice 创建设备
	CreateDevice(gid definition.DeviceGidType, name string) IDevice
	// CreateValParseObj 创建解析对象
	CreateValParseObj(params *model.ValParseParams) interface{}
}

// IDevice 设备接口
type IDevice interface {
	// Open 打开通道，等待发送指令
	Open(chanInfo model.ChannelInfo, packets model.ListCollectPackets) model2.Quality
	// Close 关闭通道
	Close() model2.Quality
	// Request 发送采集指令，并根据响应将解析后的数据填充到测点值
	Request(ctx context.Context, packet *model.CollectProtocolPacket) (model2.Quality, model3.MessageStatistics)
	// RequestPing 发送采集指令，最小化指令发送包
	RequestPing(ctx context.Context, packet model.CollectProtocolPacket) model2.Quality
	// Control 发送控制指令
	Control(packet *model.ControlProtocolPacket, val string) model2.Quality
}
