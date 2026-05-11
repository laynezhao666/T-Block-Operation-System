package iec104

import (
	"agent/entity/consts"
	"agent/entity/definition"
	"agent/logic/collector/device/driver"
	model3 "agent/logic/collector/device/model"
)

const (
	IEC104 = "iec104"
)

func init() {
	if err := driver.Register(IEC104, iec104Driver{}); err != nil {
		panic(err)
	}
}

type iec104Driver struct {
}

// Init init
func (s iec104Driver) Init() consts.Quality {
	return consts.QualityOk
}

// UnInit 反初始化
func (s iec104Driver) UnInit() consts.Quality {
	return consts.QualityOk
}

// CreateDevice 创建驱动
func (s iec104Driver) CreateDevice(gid definition.DeviceGidType, name string) driver.IDevice {
	return NewIEC104Device()
}

// CreateValParseObj 创建值解析器
func (s iec104Driver) CreateValParseObj(params *model3.ValParseParams) interface{} {
	return NewIEC104ValueParser(params)
}
