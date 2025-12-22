package main

import (
	"agent/entity/definition"
	"agent/logic/task"
	"agent/utils"
	"context"
	"etrpc-go"
	"os"
	"runtime"
	"sync"

	"trpc.group/trpc-go/trpc-go/log"

	"agent/entity/config"
	pluginInit "agent/logic/plugin/init"
	"agent/logic/setup"

	"agent/service"

	pb "trpcprotocol/agent"

	"trpc.group/trpc-go/trpc-go/restful"
)

func init() {
	// 为restful接口注册自定义序列化方式，否则响应的默认值会被丢弃
	restful.RegisterSerializer(utils.RestfulSerializer{})
	restful.SetDefaultSerializer(utils.RestfulSerializer{})
	restful.Marshaller.UseProtoNames = true
}

func main() {
	defer func() {
		if r := recover(); r != nil {
			stack := make([]byte, 102400)
			length := runtime.Stack(stack, true)
			log.Errorf("panic:%v,stack:%s",
				r, string(stack[:length]))
		}
	}()
	//s := etrpc.NewServer()
	s := etrpc.NewServer()

	var err error
	var wg sync.WaitGroup

	ctx, cancel := context.WithCancel(context.Background())

	if err = setup.Init(); err != nil {
		log.Error(err)
		os.Exit(-1)
	}
	go pluginInit.Start(ctx, &wg)

	if config.GetRB().Task.Mode == definition.TaskModeSchedule {
		go task.ReportHeartbeat(ctx, &wg)
	}

	pb.RegisterAgentCgiService(s, &service.CgiServiceImpl{})
	pb.RegisterTaskConfigService(s, &service.ScheduleServiceImpl{})
	pb.RegisterConfigManagerService(s, &service.ConfigManager{})
	pb.RegisterBoxManagerService(s, &service.BoxManager{})
	pb.RegisterRealTimeDataManagerService(s, &service.RealTimeDataServiceImpl{})
	s.RegisterOnShutdown(func() {
		log.Info("shutdown...")
		cancel()
		setup.UnInit()
	})
	etrpc.RunServer(s)

	wg.Wait()
}
