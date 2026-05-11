package kbt

import (
	"agent/entity/consts"
	"agent/entity/definition"
	"agent/logic/collector/device/driver"
	"agent/logic/collector/device/model"
	"strconv"
)

// KBTDriver 实现 IDriver
type KBTDriver struct{}

const Name = "kbt"

func init() {
	if err := driver.Register(Name, KBTDriver{}); err != nil {
		panic(err)
	}
}

func (d KBTDriver) Init() consts.Quality   { return consts.QualityOk }
func (d KBTDriver) UnInit() consts.Quality { return consts.QualityOk }

func (d KBTDriver) CreateDevice(gid definition.DeviceGidType, name string) driver.IDevice {
	return &Device{gid: gid, name: name}
}

func (d KBTDriver) CreateValParseObj(params *model.ValParseParams) interface{} {
	// KBT协议解析器参数格式：DataAddr=线路编号(1-64)
	lineNum, err := strconv.Atoi(params.DataAddr)
	if err != nil {
		return nil
	}

	return &ValueParser{
		LineNum:  uint32(lineNum),
		Extend:   params.Extend,
		DataType: params.DataType,
	}
}
