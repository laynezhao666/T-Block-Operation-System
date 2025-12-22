// Package rtdb 测点实时数据库
package rtdb

import (
	"agent/entity/definition"
	"agent/logic/collector/rtdb/model"
	"agent/utils/osal"
)

// GetPv 获取测点值
func GetPv(id definition.DataPointIDType) (osal.Variant, bool) {
	return rtdbInstance.GetPv(id)
}

// GetVal 获取测点实时值
func GetVal(id definition.DataPointIDType) (model.RTValue, bool) {
	return rtdbInstance.GetVal(id)
}

// SetVal 设置 id 对应的测点的实时值为 val；
// valChanged 如果不为 nil，如果 val 与原始测点值不同，valChanged 被置为 true，否则被置为 false；
// valChanged 如果为 nil 则忽略。
func SetVal(id definition.DataPointIDType, val *model.RTValue, valChanged *bool) {
	rtdbInstance.SetVal(id, val, valChanged)
}

// GetAlarm 获取测点的告警状态
func GetAlarm(id definition.DataPointIDType) (model.RTAlarm, bool) {
	return rtdbInstance.GetAlarm(id)
}

// SetAlarm 设置 id 对应的测点的告警状态为 alarm
func SetAlarm(id definition.DataPointIDType, alarm model.RTAlarm) {
	rtdbInstance.SetAlarm(id, alarm)
}

// GetDataPoints 根据 points 中的 id 获取对应数据并写入到数据字段
func GetDataPoints(points model.DataPoints) {
	rtdbInstance.GetDataPoints(points)
}

// GetVirtualDataPoints 获取虚拟测点数据
func GetVirtualDataPoints(points model.DataPoints) {
	rtdbInstance.GetVirtualDataPoints(points)
}

// GetDataPointsByID 根据 pointIDs 中的 id 获取对应数据并写入到数据字段
func GetDataPointsByID(pointIDs definition.DataPointIDsType) model.DataPoints {
	points := make(model.DataPoints, len(pointIDs))
	for i, id := range pointIDs {
		points[i].ID = id
	}
	GetDataPoints(points)
	return points
}

// SetDataPoints 将 points 的所有数据放入实时数据库，
// 如果 needNotify 为 true 则通知已经注册的回调，否则不通知。
func SetDataPoints(points model.DataPoints, needNotify bool) {
	rtdbInstance.SetDataPoints(points, needNotify)
}

// GetAll 获取所有测点数据
func GetAll() model.DataPoints {
	return rtdbInstance.GetAll()
}

// ClearAll 清除所有测点数据
func ClearAll() {
	rtdbInstance.ClearAll()
}

// ClearDataPoints 清除 id 在 ids 中的所有测点数据
func ClearDataPoints(ids definition.DataPointIDsType) {
	rtdbInstance.ClearDataPoints(ids)
}

// RegisterDataPointsUpdatedFun 注册回调，
// 在每次通知回调时会调用 fun，fun 的参数 arg 为调用 RegisterDataPointsUpdatedFun 的 arg
func RegisterDataPointsUpdatedFun(fun DataPointsUpdatedFun, arg interface{}) {
	rtdbInstance.RegisterDataPointsUpdatedFun(fun, arg)
}

// UnRegisterDataPointsUpdatedFun 取消 fun 的注册，
// 通知回调时不再调用 fun
func UnRegisterDataPointsUpdatedFun(fun DataPointsUpdatedFun) {
	rtdbInstance.UnregisterDataPointsUpdatedFun(fun)
}
