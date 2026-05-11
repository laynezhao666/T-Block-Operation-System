package iec103_siemens

import (
	"agent/logic/collector/device/model"
	"strconv"

	"trpc.group/trpc-go/trpc-go/log"
)

const (
	retry               = "retry"
	writeRetries        = "write_retries"
	heartbeatInterval   = "heartbeat_interval"
	maxHeartbeatTimeout = "max_heartbeat_timeout"
	totalCallInterval   = "total_call_interval"
	elecCallInterval    = "elec_call_interval"
	maxPacketSize       = "max_packet_size"
	clockSyncIntvl      = "clock_sync_intvl"
	summerTime          = "summer_time"
	maxSendFailures     = "max_send_failures"
	expiredBuf          = "expired_buf"          // 数据过期时间在采集间隔上额外加的buf
	timezone            = "timezone"             // 时区配置
	controlRespTimeout  = "control_resp_timeout" // 控制命令响应超时时间

	defaultOnceRequestTimeoutMs = 4000
	defaultTotalTimeout         = 30000 // 采集并发总超时
	defaultWriteRetries         = 1
	defaultHeartbeatInterval    = 10000
	defaultMaxHeartbeatInterval = 25000
	defaultMaxHeartbeatTimeout  = 30000
	//defaultTotalCallInterval    = 900000 // 总召唤间隔

	defaultTotalCallInterval = 660000

	defaultElecCallInterval   = 60000 // 电度召唤间隔
	defaultMaxPacketSize      = 4096
	defaultMaxSendFailures    = 5 // 默认最大发送失败次数
	minClockSyncIntvl         = 300000
	defaultClockSyncIntvl     = 600000 // 西门子建议值，10分钟
	maxClockSyncIntvl         = 900000 // 最大值，不要超过15分钟
	defaultExpiredBuf         = 10000
	defaultTimezone           = 8     // 默认时区为东八区（+8小时）
	defaultControlRespTimeout = 10000 // 默认控制命令响应超时时间，单位毫秒
)

// Option 驱动选项
type Option struct {
	ReadTimeOut         int    // 单次命令请求超时
	ReadRetries         int    // 读命令 重试次数
	WriteRetries        int    // 写命令 重试次数
	TotalTimeOut        int    // Request接口总请求超时
	HeartbeatIntvl      int    // 心跳间隔
	HeartbeatTimeout    int    //	心跳超时
	MaxHeartbeatTimeout int    // 最大心跳超时
	TotalCallInterval   int    // 总召唤间隔
	ElecCallInterval    int    // 电度召唤间隔
	MaxPacketSize       uint32 // 最大包大小
	ClockSyncIntvl      int    // 时钟同步间隔
	SummerTime          int    // 夏令时
	MaxSendFailures     int    // 最大发送失败次数（连续失败达到此值触发重连）
	ExpiredBuf          int    // 数据过期时间在采集间隔上额外加的buf
	Timezone            int    // 时区配置，表示时区偏移小时数，如+8表示东八区，-5表示西五区，0表示使用系统时区
	ControlRespTimeout  int    // 控制命令响应超时时间，单位毫秒
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
	}
	if o.ReadTimeOut < defaultOnceRequestTimeoutMs {
		o.ReadTimeOut = defaultOnceRequestTimeoutMs
	}

	v, ok := chanInfo.ExtendKV[retry]
	if ok {
		retryValue, err := strconv.Atoi(v)
		if err == nil {
			o.ReadRetries = retryValue
		}
	}
	o.WriteRetries = defaultWriteRetries
	v, ok = chanInfo.ExtendKV[writeRetries]
	if ok {
		writeRetriesValue, err := strconv.Atoi(v)
		if err == nil {
			o.WriteRetries = writeRetriesValue
		}
	}
	o.HeartbeatIntvl = defaultHeartbeatInterval
	v, ok = chanInfo.ExtendKV[heartbeatInterval]
	if ok {
		heartbeatIntervalValue, err := strconv.Atoi(v)
		if err == nil {
			o.HeartbeatIntvl = heartbeatIntervalValue
		}
	}
	if o.HeartbeatIntvl > defaultMaxHeartbeatInterval {
		o.HeartbeatIntvl = defaultMaxHeartbeatInterval
	}
	o.MaxHeartbeatTimeout = defaultMaxHeartbeatTimeout
	v, ok = chanInfo.ExtendKV[maxHeartbeatTimeout]
	if ok {
		maxHeartbeatTimeoutValue, err := strconv.Atoi(v)
		if err == nil {
			o.MaxHeartbeatTimeout = maxHeartbeatTimeoutValue
		}
	}
	if o.MaxHeartbeatTimeout < o.HeartbeatIntvl {
		o.MaxHeartbeatTimeout = o.HeartbeatIntvl + 5000
	}
	o.TotalCallInterval = defaultTotalCallInterval
	v, ok = chanInfo.ExtendKV[totalCallInterval]
	if ok {
		totalCallIntervalValue, err := strconv.Atoi(v)
		if err == nil {
			o.TotalCallInterval = totalCallIntervalValue
		}
	}
	o.ElecCallInterval = defaultElecCallInterval
	v, ok = chanInfo.ExtendKV[elecCallInterval]
	if ok {
		elecCallIntervalValue, err := strconv.Atoi(v)
		if err == nil {
			o.ElecCallInterval = elecCallIntervalValue
		}
	}
	o.MaxPacketSize = defaultMaxPacketSize
	v, ok = chanInfo.ExtendKV[maxPacketSize]
	if ok {
		maxPacketSizeValue, err := strconv.Atoi(v)
		if err == nil {
			o.MaxPacketSize = uint32(maxPacketSizeValue)
		}
	}
	v, ok = chanInfo.ExtendKV[clockSyncIntvl]
	if ok {
		clockSyncIntvlValue, err := strconv.Atoi(v)
		if err == nil {
			o.ClockSyncIntvl = clockSyncIntvlValue
		}
	}
	// 设置了时钟同步间隔则判断间隔是否在范围内，不在范围内的设置为默认值
	if o.ClockSyncIntvl > 0 && (o.ClockSyncIntvl < minClockSyncIntvl || o.ClockSyncIntvl >= maxClockSyncIntvl) {
		o.ClockSyncIntvl = defaultClockSyncIntvl
	}
	v, ok = chanInfo.ExtendKV[summerTime]
	if ok {
		summerTimeValue, err := strconv.Atoi(v)
		if err == nil {
			o.SummerTime = summerTimeValue
		}
	}

	// 设置最大发送失败次数
	o.MaxSendFailures = defaultMaxSendFailures
	v, ok = chanInfo.ExtendKV[maxSendFailures]
	if ok {
		maxSendFailuresValue, err := strconv.Atoi(v)
		if err == nil {
			o.MaxSendFailures = maxSendFailuresValue
		}
	}
	o.ExpiredBuf = defaultExpiredBuf
	v, ok = chanInfo.ExtendKV[expiredBuf]
	if ok {
		expiredBufValue, err := strconv.Atoi(v)
		if err == nil {
			o.ExpiredBuf = expiredBufValue
		}
	}

	// 设置时区配置，默认为东八区
	o.Timezone = defaultTimezone
	v, ok = chanInfo.ExtendKV[timezone]
	if ok {
		timezoneValue, err := strconv.Atoi(v)
		if err == nil {
			o.Timezone = timezoneValue
		} else {
			log.Warnf("invalid timezone value: %s, using default: %d", v, defaultTimezone)
		}
	}

	// 设置控制命令响应超时时间
	o.ControlRespTimeout = defaultControlRespTimeout
	v, ok = chanInfo.ExtendKV[controlRespTimeout]
	if ok {
		controlRespTimeoutValue, err := strconv.Atoi(v)
		if err == nil && controlRespTimeoutValue > 0 {
			o.ControlRespTimeout = controlRespTimeoutValue
		}
	}
}
