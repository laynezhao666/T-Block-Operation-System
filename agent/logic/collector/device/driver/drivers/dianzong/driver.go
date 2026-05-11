package dianzong

import (
	"agent/entity/consts"
	"agent/entity/definition"
	"agent/logic/collector/device/driver"
	"agent/logic/collector/device/model"
)

// DianzongDriver 实现 IDriver
type DianzongDriver struct{}

const (
	Dianzong = "dianzong"
)

func init() {
	if err := driver.Register(Dianzong, DianzongDriver{}); err != nil {
		panic(err)
	}
}

// Init 初始化驱动（可做全局资源准备）
func (d DianzongDriver) Init() consts.Quality {
	return consts.QualityOk
}

// UnInit 反初始化驱动
func (d DianzongDriver) UnInit() consts.Quality {
	return consts.QualityOk
}

// CreateDevice 创建设备
func (d DianzongDriver) CreateDevice(gid definition.DeviceGidType, name string) driver.IDevice {
	return &Device{
		gid:  gid,
		name: name,
	}
}

// CreateValParseObj 创建解析对象
func (d DianzongDriver) CreateValParseObj(params *model.ValParseParams) interface{} {
	// 把 params 里的扩展信息带入 ValueParser
	return &ValueParser{
		Addr:      params.DataAddr,
		Extend:    params.Extend,
		DataType:  params.DataType,
		ByteOrder: params.ByteOrder,
	}
}
