package snmp

import (
	"agent/entity/config"
	"agent/entity/consts"
	"agent/logic/collector/device/model"
	utils2 "agent/utils"
)

const (
	defaultSnmpPort            = 161
	defaultReadTimeout         = 2000
	defaultRetry               = 0
	defaultTotalTimeout        = 30000 // 并发采集时任务总超时
	defaultMaxDriverCoroutines = 100   // 默认最大并发协程数
	defaultPerCoroutinePoints  = 3000  // 默认每个协程请求的测点数

	DefaultMaxOidCount = 60
)

// Option 驱动配置
type Option struct {
	ReadTimeOut         int // SNMP Get超时
	ReadRetries         int // SNMP read 重试次数
	TotalTimeOut        int // Request接口总请求超时
	ParallelCount       int // 并发协程数量
	MaxDriverCoroutines int // 最大并发协程数
	PerCoroutinePoints  int // 每个协程请求的测点数
	PacketOIDs          int // 每次请求的oid个数
	ReadCommunity       string
	WriteCommunity      string
}

// Load 加载配置
func (o *Option) Load(chanInfo model.ChannelInfo, packets model.ListCollectPackets) {
	o.loadFromDriverConfigOrDefault()
	o.setFromChannelInfo(chanInfo, packets)
}

// loadFromDriverConfigOrDefault 从全局配置中读取，如果配置不存在，使用默认值
func (o *Option) loadFromDriverConfigOrDefault() {
	o.ReadTimeOut = config.LoadIntOrDefault(config.GetRB().Collector.Snmp.Timeout, defaultReadTimeout)
	o.ReadRetries = config.LoadIntOrDefault(config.GetRB().Collector.Snmp.Retry, defaultRetry)
	o.TotalTimeOut = config.LoadIntOrDefault(config.GetRB().Collector.Snmp.RequestTotalTimeout, defaultTotalTimeout)
	o.MaxDriverCoroutines = config.LoadIntOrDefault(config.GetRB().Collector.Snmp.MaxCoroutine, defaultMaxDriverCoroutines)
	o.PerCoroutinePoints = config.LoadIntOrDefault(config.GetRB().Collector.Snmp.PerCoroutinePoints,
		defaultPerCoroutinePoints)
}

// setFromChannelInfo 从通道参数中获取，如果通道参数中有设定值，并替换掉全局配置的值 （优先使用通道参数配置）
func (o *Option) setFromChannelInfo(chanInfo model.ChannelInfo, packets model.ListCollectPackets) {
	if chanInfo.TimeoutMs != 0 {
		o.ReadTimeOut = chanInfo.TimeoutMs
	}
	if chanInfo.PacketMaxPointCount != 0 {
		o.PacketOIDs = chanInfo.PacketMaxPointCount
	} else {
		o.PacketOIDs = DefaultMaxOidCount
	}

	if chanInfo.ParallelCount != 0 {
		o.ParallelCount = chanInfo.ParallelCount
	} else { // 根据测点数量生成并发数
		var maxPacketPoints = 0
		for i := range packets {
			if len(packets[i].Points) > maxPacketPoints {
				maxPacketPoints = len(packets[i].Points)
			}
		}

		o.ParallelCount = 1
		if maxPacketPoints%o.PerCoroutinePoints == 0 {
			o.ParallelCount = maxPacketPoints / o.PerCoroutinePoints
		} else {
			o.ParallelCount = maxPacketPoints/o.PerCoroutinePoints + 1
		}
		if o.ParallelCount > o.MaxDriverCoroutines {
			o.ParallelCount = o.MaxDriverCoroutines
		}
	}

	value := utils2.ParseKvString(chanInfo.Params, consts.SepKV)
	if len(value["read"]) > 0 {
		o.ReadCommunity = value["read"]
	} else {
		o.ReadCommunity = "public"
	}
	if len(value["write"]) > 0 {
		o.WriteCommunity = value["write"]
	} else {
		o.WriteCommunity = "private"
	}

}
