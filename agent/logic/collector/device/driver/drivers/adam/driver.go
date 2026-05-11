package adam

import (
	"agent/entity/consts"
	"agent/entity/definition"
	"agent/logic/collector/device/driver"
	"agent/logic/collector/device/model"
	"strconv"
)

// AdamDriver 实现 IDriver
type AdamDriver struct{}

const Name = "adam"

func init() {
	if err := driver.Register(Name, AdamDriver{}); err != nil {
		panic(err)
	}
}

func (d AdamDriver) Init() consts.Quality   { return consts.QualityOk }
func (d AdamDriver) UnInit() consts.Quality { return consts.QualityOk }

func (d AdamDriver) CreateDevice(gid definition.DeviceGidType, name string) driver.IDevice {
	return &Device{gid: gid, name: name}
}

func (d AdamDriver) CreateValParseObj(params *model.ValParseParams) interface{} {
	// 举例：DataAddr=0，DataType=BOOL1
	addr, err := strconv.Atoi(params.DataAddr)
	if err != nil {
		return nil
	}

	return &ValueParser{
		Addr:     uint32(addr),
		Extend:   params.Extend,
		DataType: params.DataType,
	}
}
