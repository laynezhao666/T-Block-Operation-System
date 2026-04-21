// Package cacs 实现CACS门禁控制器协议的驱动层。
package cacs

import (
	"dac/entity/consts"
	"dac/entity/model/db"
	model "dac/entity/model/driver"
	"dac/logic/collect/driver"
)

// init 注册CACS协议驱动到全局驱动管理器
func init() {
	d := cacsDriver{}
	err := driver.Register("cacs", d)
	if err != nil {
		panic(err)
	}
	if err = driver.Register("CACS", d); err != nil {
		panic(err)
	}
}

// cacsDriver CACS协议驱动工厂
type cacsDriver struct {
}

// Init 初始化CACS驱动（当前无需额外初始化）
func (h cacsDriver) Init() consts.Quality {
	return consts.QualityUncertain
}

// UnInit 反初始化CACS驱动（当前无需额外清理）
func (h cacsDriver) UnInit() consts.Quality {
	return consts.QualityUncertain
}

// CreateController 创建CACS门控器实例
func (h cacsDriver) CreateController(id db.IDType, name string) model.Controller {
	return &Controller{
		baseInfo: model.NewControllerBasicInfo(id, name),
	}
}
