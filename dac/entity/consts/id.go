// Package consts 定义门禁系统的全局常量。
package consts

// 标准测点ID常量，用于生成各类测点的唯一标识
const (
	StandardIDDoorState           = "DoorState"       // 门状态测点ID
	StandardIDCommunicationState  = "Comm"            // 通讯状态测点ID
	StandardIDFaultStatus         = "DoorFault"       // 门故障状态测点ID
	StandardIDOpenAlarm           = "DoorOpenAlarm"   // 门开超时告警测点ID
	InternalIDTotalResponseTime   = "total_resp_time" // 一个周期内的总响应时间
	IntervalIDSuccessRequestCount = "success_req"     // 成功请求计数测点ID
	IntervalIDTotalRequestCount   = "total_req"       // 总请求计数测点ID
)
