package ping

import (
	"agent/entity/consts"
	"agent/entity/definition"
	"agent/logic/collector/device/driver"
	"agent/logic/collector/device/model"
)

// PingDriver 实现 IDriver 接口
type PingDriver struct{}

// 驱动名称常量
const (
	PING = "ping"
)

func init() {
	if err := driver.Register(PING, PingDriver{}); err != nil {
		panic(err)
	}
}

// Init 初始化驱动
func (d PingDriver) Init() consts.Quality {
	return consts.QualityOk
}

// UnInit 反初始化驱动
func (d PingDriver) UnInit() consts.Quality {
	return consts.QualityOk
}

// CreateDevice 创建设备
func (d PingDriver) CreateDevice(gid definition.DeviceGidType, name string) driver.IDevice {
	return NewPingDevice(gid, name)
}

// CreateValParseObj 创建值解析对象
func (d PingDriver) CreateValParseObj(params *model.ValParseParams) interface{} {
	return NewPingValueParser(params)
}
