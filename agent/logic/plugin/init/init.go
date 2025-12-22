// Package init 初始化插件
package init

import (
	"context"
	"agent/logic/plugin/expressioncmd"
	"agent/logic/plugin/southdevice"
	consts "agent/logic/plugin/var"
	"sync"

	"trpc.group/trpc-go/trpc-go/log"

	"agent/entity/config"
	"agent/logic/plugin"

	"github.com/robfig/cron/v3"
)

var (
	c *cron.Cron
)

// Init 初始化插件
func Init() {
	consts.InterruptionJudgeThreshold = config.LoadIntOrDefault(config.GetRB().Plugin.InterruptionJudgeThreshold,
		consts.DefaultInterruptionJudgeThreshold)

	if consts.InterruptionJudgeThreshold <= 0 || consts.InterruptionJudgeThreshold > 100 {
		consts.InterruptionJudgeThreshold = 100
	}

	//注册计算带有表达式的采集点的插件
	expressioncmd.Init()
	// 注册计算南向设备通讯状态的插件
	southdevice.Init()
}

// Start 启动插件调度器
func Start(ctx context.Context, wg *sync.WaitGroup) {
	defer func() {
		if r := recover(); r != nil {
			log.Errorf("panic in plugin init: %v", r)
		}
	}()

	wg.Add(1)
	defer wg.Done()

	Init()

	// 初始化插件管理器
	plugin.Manager().Start(ctx)
	log.Info("所有插件调度器已启动")
}
