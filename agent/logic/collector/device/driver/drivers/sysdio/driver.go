package sysdio

import (
	"agent/entity/consts"
	"agent/entity/definition"
	"agent/logic/collector/device/driver"
	model3 "agent/logic/collector/device/model"
)

const (
	Sysdio = "sysdio"
)

func init() {
	if err := driver.Register(Sysdio, sysdioDriver{}); err != nil {
		panic(err)
	}
}

type sysdioDriver struct {
}

// Init 初始化
func (s sysdioDriver) Init() consts.Quality {
	return consts.QualityOk
}

// UnInit 释放资源
func (s sysdioDriver) UnInit() consts.Quality {
	return consts.QualityOk
}

// CreateDevice 创建设备
func (s sysdioDriver) CreateDevice(gid definition.DeviceGidType, name string) driver.IDevice {
	return NewSysdioDevice(gid, name)
}

// CreateValParseObj 创建解析对象
func (s sysdioDriver) CreateValParseObj(params *model3.ValParseParams) interface{} {
	return NewSysdioValueParser(params)
}
