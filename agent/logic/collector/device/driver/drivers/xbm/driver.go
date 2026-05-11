package xbm

import (
	"agent/entity/consts"
	"agent/entity/definition"
	"agent/logic/collector/device/driver"
	model1 "agent/logic/collector/device/model"
)

const xbmDriverName = "xbm"

func init() {
	if err := driver.Register(xbmDriverName, xbmDriver{}); err != nil {
		panic(err)
	}
}

type xbmDriver struct{}

// Init 初始化驱动
func (d xbmDriver) Init() consts.Quality {
	return consts.QualityOk
}

// UnInit 反初始化驱动
func (d xbmDriver) UnInit() consts.Quality {
	return consts.QualityOk
}

// CreateDevice 创建设备
func (d xbmDriver) CreateDevice(gid definition.DeviceGidType, name string) driver.IDevice {
	return NewXBMDevice()
}

// CreateValParseObj 创建值解析器
func (d xbmDriver) CreateValParseObj(params *model1.ValParseParams) interface{} {
	return NewXBMValParser(params)
}
