package main

import (
	"context"
	"fmt"
	"time"

	etrpc "etrpc-go"

	"dac/entity/config"
	"dac/entity/server/cgi"
	"dac/logic/setup"

	"trpc.group/trpc-go/trpc-go/log"
)

func main() {
	// 1. 创建 etrpc server（内部完成配置加载、插件初始化）
	s := etrpc.NewServer()

	// 2. 注册 CGI HTTP 路由
	cgi.Register(s.Service("dac"))

	// 3. 业务初始化
	config.Log = log.GetDefaultLogger()
	ctx, cancel := context.WithCancel(context.Background())
	if err := setup.Init(ctx); err != nil {
		fmt.Printf("setup init error: %v\n", err)
		config.Log.Errorf("init error: %v.", err)
		return
	}

	// 4. 注册退出回调 & 启动服务
	s.RegisterOnShutdown(func() {
		config.Log.Infof("shutdown...")
		cancel()

		exitCtx, exitCancel := context.WithTimeout(context.Background(), time.Second*10)
		defer exitCancel()

		setup.UnInit(exitCtx)
	})

	if err := s.Serve(); err != nil {
		config.Log.Errorf("start server error: %v", err)
		return
	}

	config.Log.Infof("exit now...")
}
