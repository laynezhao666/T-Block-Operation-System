package main

import (
	"context"
	"sync"

	"etrpc-go"

	"trpc.group/trpc-go/trpc-database/kafka"
	"trpc.group/trpc-go/trpc-database/timer"
	"trpc.group/trpc-go/trpc-go"

	"alarm-server/logic/cache"
	"alarm-server/logic/collector"

	"alarm-server/logic/consumer"
	"alarm-server/service/api"

	pb "trpcprotocol/alarm-server"
)

func main() {
	s := etrpc.NewServer()
	ctx, cancel := context.WithCancel(trpc.BackgroundContext())
	s.RegisterOnShutdown(func() {
		cancel()
	})
	wg := sync.WaitGroup{}
	go cache.RegularSyncCache(ctx, &wg)
	go collector.GetValidateColleror().RegularStoreValidMsg(ctx, &wg)
	kafka.RegisterBatchHandlerService(s, consumer.BatchHandleMessage)
	timer.RegisterHandlerService(
		s.Service("trpc.timer.tbos.alarmValid"), collector.ReportValidEfficiency)
	timer.RegisterHandlerService(
		s.Service("trpc.timer.tbos.delRuleRecord"), collector.ExpireInvalidRuleRecord)
	pb.RegisterAlarmServerService(s, api.NewServiceApi())
	etrpc.RunServer(s)
	wg.Wait()
}
