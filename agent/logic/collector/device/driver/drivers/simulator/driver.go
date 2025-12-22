package simulator

import (
	"agent/entity/consts"
	"agent/entity/definition"
	"agent/entity/model"
	"agent/logic/collector/device/driver"
	model3 "agent/logic/collector/device/model"
	"agent/utils"
)

func init() {
	if err := driver.Register(consts.Simulator, simulatorDriver{}); err != nil {
		panic(err)
	}
}

type simulatorDriver struct {
}

// Init 初始化
func (d simulatorDriver) Init() consts.Quality {
	return consts.QualityOk
}

// UnInit 释放资源
func (d simulatorDriver) UnInit() consts.Quality {
	return consts.QualityOk
}

// CreateDevice 创建模拟设备
func (d simulatorDriver) CreateDevice(gid definition.DeviceGidType, name string) driver.IDevice {
	device := &simulatorDevice{
		data: model.IDeviceData{
			Gid:  gid,
			Name: name,
		},
	}
	return device
}

// CreateValParseObj 创建模拟测点解析对象
func (d simulatorDriver) CreateValParseObj(params *model3.ValParseParams) interface{} {
	var bitBegin, bitEnd uint8
	p := &ValueParser{
		DataType:  utils.GetDataType(params.DataType, &bitBegin, &bitEnd),
		Generator: CreateGenerator(params.DataAddr),
	}

	return p
}
