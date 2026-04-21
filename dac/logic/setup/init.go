// Package setup 提供门禁服务的初始化和清理流程。
package setup

import (
	"context"
	"fmt"
	"os"
	"time"

	"dac/entity/config"
	"dac/entity/consts"
	"dac/entity/location"
	"dac/entity/redis"
	"dac/logic/cache"
	"dac/logic/card"
	"dac/logic/collect/controller/virtualpoints"
	"dac/logic/collect/delta"
	"dac/logic/controller"
	"dac/logic/dlm"
	"dac/logic/mapping"
	"dac/logic/push"
	"dac/logic/request"
	"dac/repo/dac"
	"dac/repo/data"

	// 注册驱动
	_ "dac/logic/collect/driver/cacs"
	_ "dac/logic/collect/driver/chd806d4"
	_ "dac/logic/collect/driver/http"
	_ "dac/logic/collect/driver/test"
	_ "dac/logic/collect/driver/xbrother"

	"etrpc-go/util/iputil"
	"trpc.group/trpc-go/trpc-go"
)

// printMessage 输出初始化进度消息到控制台和日志
func printMessage(message string) {
	fmt.Printf("%v: %v\n", time.Now(), message)
	config.Log.Info(message)
}

// Init 初始化门禁服务的所有模块，按依赖顺序启动
func Init(ctx context.Context) error {
	var err error

	config.Init()
	printMessage("init config success.")

	if err := initCGIServiceIPAndPort(); err != nil {
		return fmt.Errorf("init cgi service ip and port error: %w", err)
	}
	printMessage("init cgi service ip and port success")

	if err = location.Init(); err != nil {
		return fmt.Errorf("init location error: %w", err)
	}
	printMessage("init location success.")

	virtualpoints.Init()
	printMessage("init virtual points config.")

	if err = redis.Init(); err != nil {
		return fmt.Errorf("init redis client error: %w", err)
	}
	printMessage("init redis client success.")

	if err = dac.GetRW().Init(); err != nil {
		return fmt.Errorf("init dac read writer error: %w", err)
	}
	printMessage("init dac read writer success.")

	if err = mapping.Init(ctx); err != nil {
		return fmt.Errorf("init mapping error: %w", err)
	}
	printMessage("init mapping code -> gid.")

	dlm.Init(ctx)
	printMessage("dlm worker has inited.")

	if err = cache.Init(ctx); err != nil {
		return fmt.Errorf("init cache error: %w", err)
	}
	printMessage("init cache success.")

	delta.Init(ctx)
	printMessage("init delta.")

	card.Init(ctx)
	printMessage("init card clean.")

	request.Init(ctx)
	printMessage("init request clean")

	controller.Init(ctx)
	printMessage("init Sync controllers from CMDB success")

	if err = push.GetWorker().Start(ctx); err != nil {
		return fmt.Errorf("start push points error: %w", err)
	}
	printMessage("start push points success.")

	// 启动TBOS缓存刷新定时任务
	data.StartRefreshTBOSCacheLoop(ctx)
	printMessage("start TBOS cache refresh loop success.")

	return nil
}

// initCGIServiceIPAndPort 初始化CGI服务的IP和端口
func initCGIServiceIPAndPort() error {
	podIP := os.Getenv("POD_IP")
	if podIP == "" {
		podIP = iputil.GetLocalIP()
	}
	consts.ServiceIP = podIP
	svcs := trpc.GlobalConfig().Server.Service
	findService := false
	for i := range svcs {
		svc := &svcs[i]
		if (*svc).Name == consts.ServiceName {
			consts.ServicePort = (*svc).Port
			findService = true
			break
		}
	}
	if !findService {
		return fmt.Errorf("service not found, service name: %s", consts.ServiceName)
	}
	return nil
}

// UnInit 清理门禁服务资源，释放分布式锁
func UnInit(ctx context.Context) {
	delta.UnInit()
	printMessage("delta has uninited.")

	dlm.UnInit(ctx)
	printMessage("dlm worker has uninited.")
}
