// Package test 提供门禁控制器的测试驱动实现，用于模拟门禁设备行为。
package test

import (
	"sync/atomic"
	"time"

	"dac/entity/config"
	"dac/entity/consts"
	"dac/entity/model/db"
	driver2 "dac/entity/model/driver"
	"dac/logic/collect/driver"
)

// resetTime 调试模式自动重置时间
const (
	resetTime = time.Minute * 5
)

// debug 调试模式开关（原子操作保证并发安全）
var (
	debug atomic.Int32
)

// EnableDebug 启用调试模式，5分钟后自动关闭
func EnableDebug() {
	debug.Store(1)

	config.Log.Info("enable debug...")
	go func() {
		time.Sleep(resetTime)

		config.Log.Info("disable debug...")
		debug.Store(0)
	}()
}

// isDebug 返回当前是否处于调试模式
func isDebug() bool {
	return debug.Load() > 0
}

// init 注册测试驱动（大小写均注册）
func init() {
	debug.Store(0)

	d := testDriver{}
	err := driver.Register("test", d)
	if err != nil {
		panic(err)
	}
	if err = driver.Register("TEST", d); err != nil {
		panic(err)
	}
}

// testDriver 测试驱动工厂
type testDriver struct {
}

// Init 初始化测试驱动（无需操作）
func (h testDriver) Init() consts.Quality {
	return consts.QualityUncertain
}

// UnInit 反初始化测试驱动（无需操作）
func (h testDriver) UnInit() consts.Quality {
	return consts.QualityUncertain
}

// CreateController 创建测试控制器实例
func (h testDriver) CreateController(id db.IDType, name string) driver2.Controller {
	return &testController{
		info: driver2.NewControllerBasicInfo(id, name),
	}
}
