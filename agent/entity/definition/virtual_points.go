package definition

import (
	"fmt"
)

const (
	AlarmCountID                   = "almcount"
	AlarmStatusID                  = "almste"
	CommunicationStatusID          = "commste"
	PointThroughputID              = "point_throughput"
	RangeResponseTimeID            = "range_resp_time"       // 一个周期内的所有响应时间的极差
	TotalResponseTimeID            = "total_resp_time"       // 一个周期内的总响应时间
	AvgResponseTimeID              = "avg_resp_time"         // 一个周期内的平均响应时间
	MaxResponseTimeID              = "max_resp_time"         // 一个周期内的最大响应时间
	MinResponseTimeID              = "min_resp_time"         // 一个周期内的最小响应时间
	OnePeriodSuccessRequestCountID = "success_req_in_period" // 一个周期内的成功请求次数
	OnePeriodTotalRequestCountID   = "total_req_in_period"   // 一个周期内的总请求次数
	SuccessRequestCountID          = "success_req"
	MinuteSuccessRequestCountID    = "minute_success_req"
	TotalRequestCountID            = "total_req"
	MinuteTotalRequestCountID      = "minute_total_req"
	InterruptionID                 = "interruption"
	TimeoutRequestCountID          = "timeout_req"
	PointTmsDelayCountID           = "tms_delay_count"      // 时间戳滞后的测点个数
	PointQuaErrorCountID           = "qua_err_count"        // 质量非OK的测点个数（降噪后）
	PointOriginQuaErrorCountID     = "qua_origin_err_count" // 质量非OK的测点个数（未处理前）

	SuccessRequestMessageCountID = "success_msg_req" // 总成功报文数
	TotalRequestMessageCountID   = "total_msg_req"   // 总报文数

	// CommID 表示通讯状态，与 InterruptionID 作用完全一致。
	// 只用于兼容物模型标识符。
	CommID = "Comm"
)

const (
	TenSecondTmsDelay   = 10
	HalfMinutesTmsDelay = 30
	OneMinutesTmsDelay  = 60
)

// GenerateAlarmCountID 生成告警次数虚拟测点ID
func GenerateAlarmCountID(deviceGiD interface{}) DataPointIDType {
	return GenerateDataPointID(deviceGiD, AlarmCountID)
}

// GenerateAlarmStatusID 生成告警状态虚拟测点ID
func GenerateAlarmStatusID(deviceGiD interface{}) DataPointIDType {
	return GenerateDataPointID(deviceGiD, AlarmStatusID)
}

// GenerateCommunicationStatusID 生成通讯状态虚拟测点ID
func GenerateCommunicationStatusID(deviceGiD interface{}) DataPointIDType {
	return GenerateDataPointID(deviceGiD, CommunicationStatusID)
}

// GeneratePointThroughputID 生成测点吞吐量虚拟测点ID
func GeneratePointThroughputID(deviceGiD interface{}) DataPointIDType {
	return GenerateDataPointID(deviceGiD, PointThroughputID)
}

// GenerateRangeResponseTimeID 生成响应时间极差虚拟测点ID
func GenerateRangeResponseTimeID(deviceGiD interface{}) DataPointIDType {
	return GenerateDataPointID(deviceGiD, RangeResponseTimeID)
}

// GenerateTotalResponseTimeID 生成总响应时间虚拟测点ID
func GenerateTotalResponseTimeID(deviceGiD interface{}) DataPointIDType {
	return GenerateDataPointID(deviceGiD, TotalResponseTimeID)
}

// GenerateAvgResponseTimeID 生成平均响应时间虚拟测点ID
func GenerateAvgResponseTimeID(deviceGiD interface{}) DataPointIDType {
	return GenerateDataPointID(deviceGiD, AvgResponseTimeID)
}

// GenerateMaxResponseTimeID 生成最大响应时间虚拟测点ID
func GenerateMaxResponseTimeID(deviceGiD interface{}) DataPointIDType {
	return GenerateDataPointID(deviceGiD, MaxResponseTimeID)
}

// GenerateMinResponseTimeID 生成最小响应时间虚拟测点ID
func GenerateMinResponseTimeID(deviceGiD interface{}) DataPointIDType {
	return GenerateDataPointID(deviceGiD, MinResponseTimeID)
}

// GenerateOnePeriodSuccessRequestCountID 生成周期内成功请求次数虚拟测点ID
func GenerateOnePeriodSuccessRequestCountID(deviceGiD interface{}) DataPointIDType {
	return GenerateDataPointID(deviceGiD, OnePeriodSuccessRequestCountID)
}

// GenerateOnePeriodTotalRequestCountID 生成周期内总请求次数虚拟测点ID
func GenerateOnePeriodTotalRequestCountID(deviceGiD interface{}) DataPointIDType {
	return GenerateDataPointID(deviceGiD, OnePeriodTotalRequestCountID)
}

// GenerateSuccessRequestCountID 生成成功请求次数虚拟测点ID
func GenerateSuccessRequestCountID(deviceGiD interface{}) DataPointIDType {
	return GenerateDataPointID(deviceGiD, SuccessRequestCountID)
}

// GenerateTotalRequestCountID 生成总请求次数虚拟测点ID
func GenerateTotalRequestCountID(deviceGiD interface{}) DataPointIDType {
	return GenerateDataPointID(deviceGiD, TotalRequestCountID)
}

// GenerateMinuteSuccessRequestCountID 生成分钟成功请求次数虚拟测点ID
func GenerateMinuteSuccessRequestCountID(deviceGiD interface{}) DataPointIDType {
	return GenerateDataPointID(deviceGiD, MinuteSuccessRequestCountID)
}

// GenerateMinuteTotalRequestCountID 生成分钟总请求次数虚拟测点ID
func GenerateMinuteTotalRequestCountID(deviceGiD interface{}) DataPointIDType {
	return GenerateDataPointID(deviceGiD, MinuteTotalRequestCountID)
}

// GenerateInterruptionID 生成中断次数虚拟测点ID
func GenerateInterruptionID(deviceGiD interface{}) DataPointIDType {
	return GenerateDataPointID(deviceGiD, InterruptionID)
}

// GenerateTimeoutRequestCountID 生成超时请求次数虚拟测点ID
func GenerateTimeoutRequestCountID(deviceGiD interface{}) DataPointIDType {
	return GenerateDataPointID(deviceGiD, TimeoutRequestCountID)
}

// GenerateChannelCommID 生成通道通信状态虚拟测点ID
func GenerateChannelCommID(deviceGiD interface{}, channelIndex int) DataPointIDType {
	return GenerateDataPointID(deviceGiD, PointIDType(fmt.Sprintf("%v_%v", CommID, channelIndex+1)))
}

// GenerateCommID 生成通信状态虚拟测点ID
func GenerateCommID(deviceGiD interface{}) DataPointIDType {
	return GenerateDataPointID(deviceGiD, CommID)
}

// GeneratePointTmsDelayCountID 生成TMS延迟次数虚拟测点ID
func GeneratePointTmsDelayCountID(deviceGiD interface{}, delay int) DataPointIDType {
	return GenerateDataPointID(deviceGiD, PointIDType(fmt.Sprintf("%v_%v", PointTmsDelayCountID, delay)))
}

// GeneratePointQuaErrorCountID 生成测点Qua错误次数虚拟测点ID
func GeneratePointQuaErrorCountID(deviceGiD interface{}) DataPointIDType {
	return GenerateDataPointID(deviceGiD, PointQuaErrorCountID)
}
// GeneratePointOriginQuaErrorCountID 生成测点原始Qua错误次数虚拟测点ID
func GeneratePointOriginQuaErrorCountID(deviceGiD interface{}) DataPointIDType {
	return GenerateDataPointID(deviceGiD, PointOriginQuaErrorCountID)
}

// GenerateSuccessRequestMessageCountID 生成成功请求消息数虚拟测点ID
func GenerateSuccessRequestMessageCountID(deviceGiD interface{}) DataPointIDType {
	return GenerateDataPointID(deviceGiD, SuccessRequestMessageCountID)
}

// GenerateTotalRequestMessageCountID 生成总请求消息数虚拟测点ID
func GenerateTotalRequestMessageCountID(deviceGiD interface{}) DataPointIDType {
	return GenerateDataPointID(deviceGiD, TotalRequestMessageCountID)
}

// GetVirtualPointsID 获取虚拟测点ID
func GetVirtualPointsID(deviceGiD DeviceGidType) DataPointIDsType {
	return DataPointIDsType{
		GeneratePointThroughputID(deviceGiD),
		GenerateRangeResponseTimeID(deviceGiD),
		GenerateAvgResponseTimeID(deviceGiD),
		GenerateMaxResponseTimeID(deviceGiD),
		GenerateMinResponseTimeID(deviceGiD),
		GenerateOnePeriodSuccessRequestCountID(deviceGiD),
		GenerateOnePeriodTotalRequestCountID(deviceGiD),
		GenerateInterruptionID(deviceGiD),
		// 业务相关测点，若测点未采集，不会上报，因此获取了多余测点无影响
		GenerateCommID(deviceGiD),
		GenerateChannelCommID(deviceGiD, 0),
		GenerateChannelCommID(deviceGiD, 1),
		GenerateMinuteSuccessRequestCountID(deviceGiD),
		GenerateMinuteTotalRequestCountID(deviceGiD),
		GeneratePointTmsDelayCountID(deviceGiD, HalfMinutesTmsDelay),
		GeneratePointTmsDelayCountID(deviceGiD, OneMinutesTmsDelay),
		GeneratePointQuaErrorCountID(deviceGiD),
		GeneratePointOriginQuaErrorCountID(deviceGiD),
		GenerateTotalRequestMessageCountID(deviceGiD),
		GenerateSuccessRequestMessageCountID(deviceGiD),
	}
}
