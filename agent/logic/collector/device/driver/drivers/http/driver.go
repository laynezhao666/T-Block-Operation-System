package http

import (
	"agent/entity/consts"
	"agent/entity/definition"
	"agent/logic/collector/device/driver"
	"agent/logic/collector/device/model"
)

// HTTPDriver 实现 IDriver
type HTTPDriver struct{}

const (
	HTTP = "http"
)

func init() {
	if err := driver.Register(HTTP, HTTPDriver{}); err != nil {
		panic(err)
	}
}

// 对应的http接口建议按照https://iwiki.woa.com/p/4016447289进行实现

func (d HTTPDriver) Init() consts.Quality {
	return consts.QualityOk
}

func (d HTTPDriver) UnInit() consts.Quality {
	return consts.QualityOk
}

func (d HTTPDriver) CreateDevice(gid definition.DeviceGidType, name string) driver.IDevice {
	return NewHTTPDevice(gid, name)
}

func (d HTTPDriver) CreateValParseObj(params *model.ValParseParams) interface{} {
	return NewHTTPValueParser(params)
}
