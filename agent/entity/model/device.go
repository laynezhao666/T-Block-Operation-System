package model

import (
	"agent/utils"
	"strings"

	"agent/entity/consts"
	"agent/entity/definition"
	"agent/logic/collector/device/model"
)

// Device 设备配置数据
type Device struct {
	Gid                   definition.DeviceGidType `json:"gid"`            // 113..
	ID                    string                   `json:"id"`             // ACM_1
	Name                  string                   `json:"name"`           // 交流电量仪_1
	TypeEn                string                   `json:"device_type_en"` // ACM
	ChData                ChannelData              `json:"channel"`        // 通道数据
	TemplateData          TemplateInfo             `json:"tpl"`            // 采集模板数据
	MozuID                int                      `json:"mozu_id"`
	Extend                string                   `json:"extend"` // 对应elvdb设备里的扩展参数
	Extends               map[string]interface{}   `json:"extends"`
	NeedReopen            bool                     // 是否需要重新打开
	SubDevices            []Device                 `json:"sub_devices"` // 虚拟子设备（非直采设备）
	StdVersion            string                   `json:"std_version"`
	DevicesVersion        string                   `json:"devices_version"`
	BelongCollectorDevice string                   `json:"belong_collector_device"`
	BelongCollectorGid    string                   `json:"belong_collector_gid"`
	CollectorType         int32                    `json:"collector_type"`
	SN                    string                   `json:"sn"`
}

// Copy 复制设备数据
func (d *Device) Copy() *Device {
	if d == nil {
		return nil
	}
	newDevice := &Device{
		Gid:          d.Gid,
		ID:           d.ID,
		Name:         d.Name,
		ChData:       d.ChData,
		TemplateData: d.TemplateData,
		MozuID:       d.MozuID,
		Extends:      d.Extends,
		NeedReopen:   d.NeedReopen,
		SubDevices:   make([]Device, len(d.SubDevices)),
	}
	for i := range newDevice.SubDevices {
		newDevice.SubDevices[i] = *d.SubDevices[i].Copy()
	}
	return newDevice
}

// GetDeviceInfo 获取对应的设备信息
func (d *Device) GetDeviceInfo() DeviceInfo {
	if d == nil {
		return DeviceInfo{}
	}

	// 配置中未设置的字段赋默认值
	cmdInterval := d.ChData.CmdInterval
	waitTimeMs := d.ChData.WaitTimeMs
	if d.ChData.CmdInterval == 0 {
		if d.ChData.Chtype == definition.ChannelTypeSerial {
			cmdInterval = consts.DefaultSerialDeviceCmdIntervalMs
		} else {
			cmdInterval = consts.DefaultNetDeviceCmdIntervalMs
		}
	}
	//else if d.ChData.CmdInterval > 200 {
	//	// 特殊兼容，原版的间隔慢300ms，故对需要特别延迟的场景加300ms
	//	cmdInterval += 300
	//}
	if d.ChData.WaitTimeMs == 0 {
		waitTimeMs = consts.DefaultChannelDeviceWaitMs
	}

	info := DeviceInfo{
		Gid:                 d.Gid,
		ID:                  d.ID,
		Name:                d.Name,
		ChannelID:           d.ChData.ChannelID,
		ChannelParams:       d.ChData.ChannelParams,
		Address:             d.ChData.Address,
		ProtocolVersion:     d.ChData.ProtocolVersion,
		Extends:             d.Extends,
		ChannelExtend:       d.Extend, // 调整对应到elvdb设备里的扩展参数
		Template:            d.TemplateData.GetFullTemplateName(),
		CmdInterval:         cmdInterval,
		WaitTimeMs:          waitTimeMs,
		TimeoutMs:           d.ChData.TimeoutMs,
		NeedReopen:          d.NeedReopen,
		ParallelCount:       d.ChData.ParallelCount,
		PacketMaxPointCount: d.ChData.PacketMaxPointCount,
		ChType:              d.ChData.Chtype,
	}
	info.FillChannels()
	info.FillExtendKV()

	return info
}

// IDeviceData 设备接口数据
type IDeviceData struct {
	// 设备 Gid
	Gid definition.DeviceGidType
	// 设备名称
	Name string
}

// DeviceInfo 设备信息
type DeviceInfo struct {
	Gid  definition.DeviceGidType `json:"gid"`
	ID   string                   `json:"id"`
	Name string                   `json:"name"`
	// 通道 Gid，为 ';' 分割的多个 IP:Port 地址或串口号，如 "192.168.1.100:502;192.168.1.101:161"
	ChannelID string `json:"chid"`
	// 各通道对应的参数，';' 分割
	ChannelParams string `json:"chparams"`
	// 各通道对应的设备地址，';' 分割
	Address string `json:"addr"`
	// 协议版本
	ProtocolVersion string `json:"prot_ver"`
	// 设备扩展参数
	Extends map[string]interface{}
	// 通道扩展参数
	ChannelExtend string `json:"extend"`
	// 模板名称
	Template string `json:"tpl"`
	// cmd命令间的等待时间
	CmdInterval int `json:"cmd_interval"`
	// 采集完成后的等待时间
	WaitTimeMs int `json:"wait_time"`
	// 采集的超时时间, 暂未使用
	TimeoutMs int `json:"timeout"`
	// 是否需要重新打开，当与通讯、协议相关的属性发生变化时必须置为 true
	NeedReopen bool
	// 并发协程数
	ParallelCount int `json:"parallel_count"`
	// 请求包允许的最大测点数
	PacketMaxPointCount int `json:"packet_max_point_count"`
	// 驱动中使用的通讯地址列表
	Channels []model.Channel
	// 扩展参数的KV结构
	ChannelExtendKV map[string]string
	// 通道类型
	ChType       string
	DriverExtend string
}

// IsMultiChannel 是否为多通道
func (d *DeviceInfo) IsMultiChannel() bool {
	if d == nil {
		return false
	}
	return len(d.Channels) > 1
}

// FillExtendKV 填充扩展参数的 KV 结构
func (d *DeviceInfo) FillExtendKV() {
	d.ChannelExtendKV = utils.ParseKvString(d.ChannelExtend, consts.SepKV)
}

// FillChannels 填充驱动中使用的通讯地址列表
func (d *DeviceInfo) FillChannels() {
	if d == nil {
		return
	}
	ids := strings.Split(d.ChannelID, consts.Sep)
	params := strings.Split(d.ChannelParams, consts.Sep)
	addresses := strings.Split(d.Address, consts.Sep)
	idLen := len(ids)
	paramLen := len(params)
	addressLen := len(addresses)
	channels := make([]model.Channel, idLen)
	for i := 0; i < idLen; i++ {
		channels[i].Name = ids[i]
		if i < paramLen {
			channels[i].Params = params[i]
		} else {
			// 若分割后的参数个数不足，自动使用上一个参数补齐
			channels[i].Params = channels[i-1].Params
		}
		if i < addressLen {
			channels[i].Address = addresses[i]
		} else {
			// 补齐
			channels[i].Address = channels[i-1].Address
		}
	}
	d.Channels = channels
}

// SubDeviceData 子设备数据
type SubDeviceData struct {
	DeviceType        definition.DeviceGidType
	DeviceGiD         definition.DeviceGidType
	InstanceDeviceGid definition.DeviceGidType
	PointsInfo        model.InstancePointsInfo
	StdPointsInfo     model.StdInstancePointsInfo
}

// MessageStatistics 报文统计
type MessageStatistics struct {
	SendCount    uint64 // 与设备通讯交互的报文条数（由具体的驱动赋值）
	SuccessCount uint64 // 与设备通讯交互成功的报文条数（由具体的驱动赋值）
	ErrLog       error  // 详细错误
	RecvPackets  string // 回包信息
}
