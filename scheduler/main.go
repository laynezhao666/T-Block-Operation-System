package main

import (
	"etrpc-go"
	"scheduler/service"
	"trpcprotocol/scheduler"
)

func main() {
	s := etrpc.NewServer()

	scheduler.RegisterRegisterService(s, service.NewRegisterService())
	scheduler.RegisterAdminService(s, service.NewAdminService())

	// 首次启动所有任务
	service.GetSchedulerService().RefreshTask()

	// 收到停止信号后，取消所有调度任务
	s.RegisterOnShutdown(func() {
		service.GetSchedulerService().CancelTask()
	})

	etrpc.RunServer(s)

	// 结束后，等待所有调度任务执行完毕
	service.GetSchedulerService().WaitTaskDone()
}
