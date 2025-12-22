package modbus

import (
	"agent/entity/consts"
	"agent/entity/definition"
	"agent/logic/collector/device/driver"
	model1 "agent/logic/collector/device/model"
)

const driverName = "modbus"

func init() {
	if err := driver.Register(driverName, modbusDriver{}); err != nil {
		panic(err)
	}
}

type modbusDriver struct {
}

// Init init
func (s modbusDriver) Init() consts.Quality {
	return consts.QualityOk
}

// UnInit 反初始化
func (s modbusDriver) UnInit() consts.Quality {
	return consts.QualityOk
}

// CreateDevice 创建驱动
func (s modbusDriver) CreateDevice(gid definition.DeviceGidType, name string) driver.IDevice {
	return NewModbusDevice()
}

// CreateValParseObj 创建值解析器
func (s modbusDriver) CreateValParseObj(params *model1.ValParseParams) interface{} {
	return NewModbusValParser(params)
}
