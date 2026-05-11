// Package dlt DL/T 645-2007 多功能电能表通信协议驱动
package dlt

import (
	"agent/entity/consts"
	"agent/entity/definition"
	"agent/logic/collector/device/driver"
	model1 "agent/logic/collector/device/model"
)

const dltDriverName = "dlt645"

func init() {
	if err := driver.Register(dltDriverName, dltDriver{}); err != nil {
		panic(err)
	}
}

type dltDriver struct{}

// Init 初始化驱动
func (d dltDriver) Init() consts.Quality {
	return consts.QualityOk
}

// UnInit 反初始化驱动
func (d dltDriver) UnInit() consts.Quality {
	return consts.QualityOk
}

// CreateDevice 创建设备
func (d dltDriver) CreateDevice(gid definition.DeviceGidType, name string) driver.IDevice {
	return NewDLTDevice()
}

// CreateValParseObj 创建值解析器
func (d dltDriver) CreateValParseObj(params *model1.ValParseParams) interface{} {
	return NewDLTValParser(params)
}
