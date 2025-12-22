package model

import (
	"agent/entity/consts"
	"agent/entity/definition"
	"agent/utils"
)

// DataPoint 测点数据
type DataPoint struct {
	// 测点 ID
	ID definition.DataPointIDType `json:"id"`
	// 设备ID
	DeviceGiD definition.DeviceGidType `json:"-"`
	// 测点数据
	Rtd RTData `json:"rtd"`
	// 当前测点值是否与上次采集的测点值不同
	IsValueChanged bool `json:"is_value_changed"`
	// 测点类型 0 采集；1 标准
	PointType definition.PointType `point_type`
}

// DataPoints 测点数据数组
type DataPoints []DataPoint

// NewVirtualDataPoint 创建虚拟测点
func NewVirtualDataPoint(deviceGiD definition.DeviceGidType, pointID definition.DataPointIDType) DataPoint {
	return DataPoint{
		DeviceGiD: deviceGiD,
		ID:        pointID,
		Rtd:       NewVirtualRTData(),
	}
}

// NewVirtualDataPointWithValue 创建虚拟测点
func NewVirtualDataPointWithValue(deviceGiD definition.DeviceGidType, pointID definition.DataPointIDType,
	value interface{}) DataPoint {
	p := NewVirtualDataPoint(deviceGiD, pointID)
	p.SetValue(value)
	return p
}

// SetValue 设置测点值
func (p *DataPoint) SetValue(value interface{}) {
	if p == nil {
		return
	}
	p.Rtd.Val.Pv.SetValue(value)
	p.Rtd.Val.Qua = consts.QualityOk
	p.Rtd.Val.Tms = utils.GetNowUTCTimeStamp()
}

// SetValueWithTime 设置测点值
func (p *DataPoint) SetValueWithTime(value interface{}, tms int64) *DataPoint {
	if p == nil {
		return nil
	}
	p.Rtd.Val.Pv.SetValue(value)
	p.Rtd.Val.Qua = consts.QualityOk
	p.Rtd.Val.Tms = tms
	return p
}

// SetValueWithTimeAndDesc 设置测点值
func (p *DataPoint) SetValueWithTimeAndDesc(value interface{}, tms int64, desc string) *DataPoint {
	if p == nil {
		return nil
	}
	p.SetValueWithTime(value, tms)
	p.Rtd.Val.Desc = desc
	return p
}
