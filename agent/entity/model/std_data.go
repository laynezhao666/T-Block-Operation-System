package model

import (
	"agent/logic/collector/device/model"
	"math"
)

// StdData 标准化数据
type StdData struct {
	// 标准化信息
	StdCommon model.StdCommon `json:"std_common"`
	// 标准测点列表
	StdPointsInfo model.StdInstancePointsInfo `json:"std_points"`
	// 子设备数据
	SubDevices []SubDeviceData
}

// Copy 复制
func (t *StdData) Copy() *StdData {
	if t == nil {
		return nil
	}
	newStdData := new(StdData)
	newStdData.StdCommon = t.StdCommon
	newStdData.StdPointsInfo = make(model.StdInstancePointsInfo, len(t.StdPointsInfo))
	newStdData.SubDevices = make([]SubDeviceData, len(t.SubDevices))

	copy(newStdData.StdPointsInfo, t.StdPointsInfo)
	for i := range newStdData.SubDevices {
		newStdData.SubDevices[i] = t.SubDevices[i]
		newStdData.SubDevices[i].PointsInfo = make(model.InstancePointsInfo, len(t.SubDevices[i].PointsInfo))
		newStdData.SubDevices[i].StdPointsInfo = make(model.StdInstancePointsInfo, len(t.SubDevices[i].StdPointsInfo))
		copy(newStdData.SubDevices[i].PointsInfo, t.SubDevices[i].PointsInfo)
		copy(newStdData.SubDevices[i].StdPointsInfo, t.SubDevices[i].StdPointsInfo)
	}
	return newStdData
}

// GetPoints 获取测点
func (t *StdData) GetPoints() model.StdInstancePointsInfo {
	if t == nil {
		return nil
	}
	points := make(model.StdInstancePointsInfo, 0, int(math.Max(float64(len(t.StdPointsInfo)), 1))*len(t.SubDevices))
	points = append(points, t.StdPointsInfo...)
	for _, sub := range t.SubDevices {
		points = append(points, sub.StdPointsInfo...)
	}
	return points
}
