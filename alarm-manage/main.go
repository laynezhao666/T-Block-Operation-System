package main

import (
	"context"
	"sync"

	pb "trpcprotocol/alarm-manage"

	"etrpc-go"

	"trpc.group/trpc-go/trpc-database/kafka"
	"trpc.group/trpc-go/trpc-go"

	"alarm-manage/logic/consumer"
	"alarm-manage/logic/lcache"
	"alarm-manage/logic/manager"
	"alarm-manage/logic/snowflake"
	"alarm-manage/service"
)

func main() {
	s := etrpc.NewServer()
	ctx, cancel := context.WithCancel(trpc.BackgroundContext())
	s.RegisterOnShutdown(func() {
		cancel()
	})
	wg := sync.WaitGroup{}
	// 雪花算法分布式Redis ID抢占
	// 若抢占失败，则影响告警全局唯一ID生成，panic
	snowflake.InitSnowflake(ctx)
	go manager.GetGlobalManager().Run(ctx, &wg)
	go lcache.GetCacheAgent().RegularSyncDevice(ctx, &wg)
	pb.RegisterManageService(s, &service.ManageService{})
	kafka.RegisterKafkaConsumerService(s, consumer.Consumer{})
	etrpc.RunServer(s)
	wg.Wait()
}
