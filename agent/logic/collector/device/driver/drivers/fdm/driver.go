package fdm

import (
	"agent/entity/consts"
	"agent/entity/definition"
	"agent/logic/collector/device/driver"
	model1 "agent/logic/collector/device/model"
)

const fdmDriverName = "fdm"

func init() {
	if err := driver.Register(fdmDriverName, fdmDriver{}); err != nil {
		panic(err)
	}
}

type fdmDriver struct{}

// Init 初始化驱动
func (d fdmDriver) Init() consts.Quality {
	return consts.QualityOk
}

// UnInit 反初始化驱动
func (d fdmDriver) UnInit() consts.Quality {
	return consts.QualityOk
}

// CreateDevice 创建设备
func (d fdmDriver) CreateDevice(gid definition.DeviceGidType, name string) driver.IDevice {
	return NewFDMDevice()
}

// CreateValParseObj 创建值解析器
func (d fdmDriver) CreateValParseObj(params *model1.ValParseParams) interface{} {
	return NewFDMValParser(params)
}
