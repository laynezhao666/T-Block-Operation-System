package model

import (
	"agent/entity/consts"
	"agent/utils/osal"
)

// RTData 测点实时告警数据
type RTData struct {
	Val     RTValue `json:"val"`
	Alarm   RTAlarm `json:"alarm"`
	Virtual bool    `json:"virtual"`
}

func newRTData(virtual bool) RTData {
	return RTData{
		Val:     NewRTValue(),
		Alarm:   NewRTAlarm(),
		Virtual: virtual,
	}
}

// NewRTData 新建一个 RTData 实例
func NewRTData() RTData {
	return newRTData(false)
}

// NewVirtualRTData 新建一个虚拟的 RTData 实例
func NewVirtualRTData() RTData {
	return newRTData(true)
}

// NewVirtualRtDataWithValueTime 新建一个虚拟的带值的 RTData 实例
func NewVirtualRtDataWithValueTime(value interface{}, time int64) RTData {
	return RTData{
		Val: RTValue{
			Pv: osal.NewVariantWithValue(value),
			// qua 固定为 ok
			Qua:  consts.QualityOk,
			Tms:  time,
			Desc: "",
		},
		Alarm:   NewRTAlarm(),
		Virtual: true,
	}
}
