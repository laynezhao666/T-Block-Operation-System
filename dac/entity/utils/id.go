// Package utils 提供门禁系统通用工具函数。
package utils

import (
	"fmt"

	"dac/entity/consts"
	"dac/entity/model/db"
)

// GeneratePointID 根据设备ID和点位ID生成复合点位标识
func GeneratePointID(deviceID, pointID interface{}) string {
	return fmt.Sprintf("%v.%v", deviceID, pointID)
}

// GenerateDoorStateID 生成门状态点位ID
func GenerateDoorStateID(controllerID db.IDType, doorNumber int) string {
	return GeneratePointID(controllerID,
		fmt.Sprintf("%v_%v", consts.StandardIDDoorState, doorNumber))
}

// GenerateDoorOpenAlarmID 生成门开超时告警点位ID
func GenerateDoorOpenAlarmID(controllerID db.IDType, doorNumber int) string {
	return GeneratePointID(controllerID,
		fmt.Sprintf("%v_%v", consts.StandardIDOpenAlarm, doorNumber))
}

// GenerateCommID 生成通讯状态点位ID
func GenerateCommID(controllerID db.IDType) string {
	return GeneratePointID(controllerID, consts.StandardIDCommunicationState)
}

// GenerateFaultID 生成故障状态点位ID
func GenerateFaultID(controllerID db.IDType) string {
	return GeneratePointID(controllerID, consts.StandardIDFaultStatus)
}

// GenerateTotalResponseTimeID 生成总响应时间点位ID
func GenerateTotalResponseTimeID(deviceID interface{}) string {
	return GeneratePointID(deviceID, consts.InternalIDTotalResponseTime)
}

// GenerateSuccessRequestCountID 生成成功请求计数点位ID
func GenerateSuccessRequestCountID(deviceID interface{}) string {
	return GeneratePointID(deviceID, consts.IntervalIDSuccessRequestCount)
}

// GenerateTotalRequestCountID 生成总请求计数点位ID
func GenerateTotalRequestCountID(deviceID interface{}) string {
	return GeneratePointID(deviceID, consts.IntervalIDTotalRequestCount)
}

// GenerateFullPointID 生成带前缀的完整Redis点位Key
func GenerateFullPointID(id string) string {
	return "t_dac::" + id
}
