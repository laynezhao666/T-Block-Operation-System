// Package chd806d4 实现CHD806D4门禁控制器协议的驱动层。
package chd806d4

import (
	"dac/entity/consts"
	"dac/entity/model/db"
	driverModel "dac/entity/model/driver"
	"dac/logic/collect/driver"
	"strings"
)

// init 注册CHD806D4驱动到驱动管理器
func init() {
	// 注册 CHD 驱动
	d := ChdDriver{}
	err := driver.Register(consts.ProtocolChd806d4, d)
	if err != nil {
		panic(err)
	}
	if err = driver.Register(strings.ToUpper(consts.ProtocolChd806d4), d); err != nil {
		panic(err)
	}
}

// ChdDriver CHD806D4 驱动
type ChdDriver struct {
}

// Init 初始化驱动（暂未实现）
func (h ChdDriver) Init() consts.Quality {
	return consts.QualityUncertain
}

// UnInit 反初始化驱动（暂未实现）
func (h ChdDriver) UnInit() consts.Quality {
	return consts.QualityUncertain
}

// CreateController 创建CHD806D4门控器实例
func (h ChdDriver) CreateController(id db.IDType, name string) driverModel.Controller {
	return &Controller{
		baseInfo: driverModel.NewControllerBasicInfo(id, name),
	}
}
