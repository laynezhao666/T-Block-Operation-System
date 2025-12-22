package model

import (
	"agent/entity/consts"
	"agent/utils"
	"agent/utils/osal"
)

// RTValue 测点实时值
type RTValue struct {
	// 测点值
	Pv osal.Variant `json:"pv"`
	// 测点质量
	Qua consts.Quality `json:"qua"`
	// 采集时间
	Tms int64 `json:"tms"`
	// 最近一次Qua=QualityOk的时间
	LastQuaOkTms int64 `json:"last_qua_ok_tms"`
	// 值的描述
	Desc string `json:"desc"`
}

// NotCollected 判断测点是否未采集
func (r *RTValue) NotCollected() bool {
	if r == nil {
		return true
	}
	return r.Qua == consts.QualityUncollected
}

// IsOK 判断测点质量是否为OK
func (r *RTValue) IsOK() bool {
	if r == nil {
		return false
	}
	return r.Qua == consts.QualityOk
}

// IsSubDeviceErr 判断测点是否为子设备错误
func (r *RTValue) IsSubDeviceErr() bool {
	if r == nil {
		return false
	}
	return r.Qua == consts.QualityUnderBoxNorthErr
}

// TmsDiffByNow 获取测点采集时间与当前时间差值
func (r *RTValue) TmsDiffByNow() int64 {
	if r == nil {
		return utils.GetNowUTCTimeStamp()
	}
	return utils.GetNowUTCTimeStamp() - r.Tms
}

// LastQuaOkTmsDiffByNow 获取测点最近一次Qua=QualityOk的时间与当前时间差值
func (r *RTValue) LastQuaOkTmsDiffByNow() int64 {
	if r == nil {
		return utils.GetNowUTCTimeStamp()
	}
	return utils.GetNowUTCTimeStamp() - r.LastQuaOkTms
}

// NewRTValue 创建一个 RTValue 实例
func NewRTValue() RTValue {
	return RTValue{
		Pv:   osal.NewVariant(),
		Qua:  consts.QualityUncollected,
		Tms:  0,
		Desc: "",
	}
}
