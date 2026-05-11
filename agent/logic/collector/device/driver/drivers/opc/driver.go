package opc

import (
	"agent/entity/consts"
	"agent/entity/definition"
	"agent/logic/collector/device/driver"
	model3 "agent/logic/collector/device/model"
)

const (
	OPCUA = "opcua"
)

func init() {
	if err := driver.Register(OPCUA, opcuaDriver{}); err != nil {
		panic(err)
	}
}

type opcuaDriver struct {
}

// Init init
func (s opcuaDriver) Init() consts.Quality {
	return consts.QualityOk
}

// UnInit 反初始化
func (s opcuaDriver) UnInit() consts.Quality {
	return consts.QualityOk
}

// CreateDevice 创建驱动
func (s opcuaDriver) CreateDevice(gid definition.DeviceGidType, name string) driver.IDevice {
	return &Device{}
}

// CreateValParseObj 创建值解析器
func (s opcuaDriver) CreateValParseObj(params *model3.ValParseParams) interface{} {
	return NewOpcuaValueParser(params)
}
