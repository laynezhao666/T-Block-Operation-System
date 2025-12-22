package main

import (
	"context"
	"sync"

	"etrpc-go"
	pb "trpcprotocol/alarm-compute"

	"trpc.group/trpc-go/trpc-go"

	"alarm-compute/logic/collector/validate"
	"alarm-compute/logic/collector/vtpoint"
	"alarm-compute/logic/rules/rmanager"

	"alarm-compute/logic/heartbeat"
	"alarm-compute/logic/strategy"
	"alarm-compute/service"
)

func main() {
	s := etrpc.NewServer()

	ctx, cancel := context.WithCancel(trpc.BackgroundContext())
	s.RegisterOnShutdown(func() {
		cancel()
	})

	wg := sync.WaitGroup{}
	// 心跳上报
	go heartbeat.ReportHeartbeat(ctx, &wg)
	// 策略接收器
	go strategy.GetStrategyHandler().Run(ctx, &wg)
	// 策略生效收集器
	go validate.GetTaskCollector().StartValidate(ctx, &wg)
	// 虚拟测点采集器，
	go vtpoint.GetPointCollector().ReportVtPointData(ctx, &wg)
	// 失败策略收集器，用于重新执行
	go validate.GetFailRuleCollector().CollectFailed(ctx, &wg)
	// 策略计算
	go run(ctx, &wg)

	pb.RegisterAlarmComputeService(s, service.NewAlarmServiceImpl())
	etrpc.RunServer(s)
	wg.Wait()
}

func run(ctx context.Context, wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()
	m := rmanager.GetGlobalRuleManager()
	go m.StartDelayTimeRuleTask(ctx)
	go m.StartVirtualRuleTask(ctx)
	m.StartRealTimeRuleTask(ctx)
}
