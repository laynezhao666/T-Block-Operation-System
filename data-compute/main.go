package main

import (
	"context"
	"data-compute/logic/compute"
	"data-compute/logic/register"
	"data-compute/service"
	"etrpc-go"
	"sync"
	"trpcprotocol/data-compute"

	"trpc.group/trpc-go/trpc-go"
)

func main() {
	// 创建 etrpc 服务器
	s := etrpc.NewServer()

	// 注册优雅关闭hook, 收到停止信号后，定时任务不再执行，等待执行中的任务执行完成
	ctx, cancel := context.WithCancel(trpc.BackgroundContext())
	s.RegisterOnShutdown(func() {
		cancel()
	})
	wg := sync.WaitGroup{}

	data_compute.RegisterComputeService(s, service.NewComputeService())
	go compute.GetComputeApi().StartCalcPoint(ctx, &wg)

	// 定时上报心跳
	go register.ReportHeartbeat(ctx, &wg)

	// 运行 etrpc 服务器
	etrpc.RunServer(s)

	wg.Wait()
}
