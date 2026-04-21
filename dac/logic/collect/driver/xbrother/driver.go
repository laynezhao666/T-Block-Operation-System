// Package xbrother 实现XBrother门禁控制器协议的驱动层。
package xbrother

import (
	"strings"

	"dac/entity/consts"
	"dac/entity/model/db"
	driverModel "dac/entity/model/driver"
	"dac/logic/collect/driver"
)

// init 注册XBrother协议驱动（大小写均注册）
func init() {
	d := Driver{}
	err := driver.Register(consts.ProtocolXBrother, d)
	if err != nil {
		panic(err)
	}
	if err = driver.Register(strings.ToUpper(consts.ProtocolXBrother), d); err != nil {
		panic(err)
	}
}

// Driver XBrother协议驱动工厂
type Driver struct {
}

// Init 初始化驱动（XBrother驱动无需初始化操作）
func (h Driver) Init() consts.Quality {
	return consts.QualityUncertain
}

// UnInit 反初始化驱动（XBrother驱动无需清理操作）
func (h Driver) UnInit() consts.Quality {
	return consts.QualityUncertain
}

// CreateController 创建XBrother协议控制器实例
func (h Driver) CreateController(id db.IDType, name string) driverModel.Controller {
	return &Controller{
		baseInfo: driverModel.NewControllerBasicInfo(id, name),
	}
}
