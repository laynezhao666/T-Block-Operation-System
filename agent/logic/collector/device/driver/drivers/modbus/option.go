package modbus

import (
	"agent/logic/collector/device/model"
	"strconv"
)

const (
	enableShareTransportParam   = "enable_share_transport"
	retry                       = "retry"
	defaultOnceRequestTimeoutMs = 2000
	defaultTotalTimeout         = 30000 // 采集并发总超时
	defaultWriteRetries         = 1
)

// Option 驱动选项
type Option struct {
	ReadTimeOut    int  // 单次modbus 命令请求超时
	ReadRetries    int  // modbus 读命令 重试次数
	WriteRetries   int  // modbus 写命令 重试次数
	TotalTimeOut   int  // Request接口总请求超时
	shareTransport bool // 同一个IP+端口+地址的设备，共用1个连接
}

// Load 从通道参数中加载配置
func (o *Option) Load(chanInfo model.ChannelInfo, packets model.ListCollectPackets) {
	o.setFromChannelInfo(chanInfo, packets)
}

// setFromChannelInfo 从通道参数中获取，如果通道参数中有设定值，并替换掉全局配置的值 （优先使用通道参数配置）
func (o *Option) setFromChannelInfo(chanInfo model.ChannelInfo, packets model.ListCollectPackets) {
	o.TotalTimeOut = defaultTotalTimeout
	if chanInfo.TimeoutMs != 0 {
		o.ReadTimeOut = chanInfo.TimeoutMs
	} else {
		o.ReadTimeOut = defaultOnceRequestTimeoutMs
	}

	v, ok := chanInfo.ExtendKV[enableShareTransportParam]
	if ok && v == "1" {
		o.shareTransport = true
	}

	v, ok = chanInfo.ExtendKV[retry]
	if ok {
		retryValue, err := strconv.Atoi(v)
		if err == nil {
			o.ReadRetries = retryValue
		}
	}
	o.WriteRetries = defaultWriteRetries
}
