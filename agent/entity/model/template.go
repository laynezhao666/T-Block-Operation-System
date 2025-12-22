package model

import (
	"math"
	"path/filepath"

	"agent/logic/collector/device/model"
)

// TemplateInfo 采集驱动模板数据
type TemplateInfo struct {
	TemplateName string `json:"tplnm"`
	TemplatePath string `json:"tplpath"`
}

// GetFullTemplateName 获取模板完整路径
func (d TemplateInfo) GetFullTemplateName() string {
	return filepath.Join(d.TemplatePath, d.TemplateName)
}

// TemplateData 模板数据
type TemplateData struct {
	// 驱动信息
	DrvInfo model.DriverInfo `json:"drvinfo"`
	// 测点信息
	PointsInfo model.InstancePointsInfo `json:"points"`
	// 子设备数据
	SubDevices []SubDeviceData
}

// Copy 复制模板数据
func (t *TemplateData) Copy() *TemplateData {
	if t == nil {
		return nil
	}
	// newTemplate := new(TemplateData)
	newTemplate := &TemplateData{
		DrvInfo: model.DriverInfo{
			Class:           t.DrvInfo.Class,
			Vendor:          t.DrvInfo.Vendor,
			DriverName:      t.DrvInfo.DriverName,
			ProtocolVersion: t.DrvInfo.ProtocolVersion,
			Extend:          t.DrvInfo.Extend,
		},
	}
	newTemplate.PointsInfo = make(model.InstancePointsInfo, len(t.PointsInfo))
	newTemplate.SubDevices = make([]SubDeviceData, len(t.SubDevices))

	copy(newTemplate.PointsInfo, t.PointsInfo)
	for i := range newTemplate.SubDevices {
		newTemplate.SubDevices[i] = t.SubDevices[i]
		newTemplate.SubDevices[i].PointsInfo = make(model.InstancePointsInfo, len(t.SubDevices[i].PointsInfo))
		copy(newTemplate.SubDevices[i].PointsInfo, t.SubDevices[i].PointsInfo)
	}
	return newTemplate
}

// GetPoints 获取模板中的所有测点
func (t *TemplateData) GetPoints() model.InstancePointsInfo {
	if t == nil {
		return nil
	}
	points := make(model.InstancePointsInfo, 0, int(math.Max(float64(len(t.PointsInfo)), 1))*len(t.SubDevices))
	points = append(points, t.PointsInfo...)
	for _, sub := range t.SubDevices {
		points = append(points, sub.PointsInfo...)
	}
	return points
}
