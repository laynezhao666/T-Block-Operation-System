// Package http 实现HTTP门禁控制器协议的驱动层。
package http

import (
	"dac/entity/consts"
	"dac/entity/model/db"
	driverModel "dac/entity/model/driver"
	"dac/logic/collect/driver"
)

// init 注册HTTP协议驱动到全局驱动管理器
func init() {
	d := httpDriver{}
	err := driver.Register("http", d)
	if err != nil {
		panic(err)
	}
	if err = driver.Register("HTTP", d); err != nil {
		panic(err)
	}
}

// httpDriver HTTP协议驱动工厂
type httpDriver struct {
}

// Init 初始化HTTP驱动（当前无需额外初始化）
func (h httpDriver) Init() consts.Quality {
	return consts.QualityUncertain
}

// UnInit 反初始化HTTP驱动（当前无需额外清理）
func (h httpDriver) UnInit() consts.Quality {
	return consts.QualityUncertain
}

// CreateController 创建HTTP门控器实例
func (h httpDriver) CreateController(id db.IDType, name string) driverModel.Controller {
	return &Controller{
		baseInfo: driverModel.NewControllerBasicInfo(id, name),
	}
}
