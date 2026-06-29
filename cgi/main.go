package main

import (
	"sync"

	"etrpc-go"

	"trpc.group/trpc-go/trpc-database/kafka"
	"trpc.group/trpc-go/trpc-go"
	thttp "trpc.group/trpc-go/trpc-go/http"
	"trpc.group/trpc-go/trpc-go/server"

	"cgi/logic"
	"cgi/logic/alarm/consumer"
	"cgi/logic/alarm/wslogic"
	"cgi/repo/cache"
	"cgi/service"

	pb "trpcprotocol/cgi"
)

func main() {
	// trpc server
	s := etrpc.NewServer(server.WithNamedFilter("mozu_filter", logic.MozuIdServerFilter))

	// default context
	ctx := trpc.BackgroundContext()

	wg := sync.WaitGroup{}

	pb.RegisterCmdbService(s, service.NewCmdbService())
	pb.RegisterDataService(s, service.NewDataService())
	pb.RegisterCommonService(s, service.NewCommonService())
	pb.RegisterAlarmService(s, service.NewAlarmService())

	kafka.RegisterBatchHandlerService(s, consumer.BatchHandleMessage)

	cache.InitCache(ctx)

	alarmWSImpl := wslogic.GetAlarmWSImpl()
	go alarmWSImpl.ExecPushAlarm(ctx, &wg)
	go alarmWSImpl.RegularPushAll(ctx, &wg)
	thttp.HandleFunc("/ws", alarmWSImpl.HandleWebSocket)
	thttp.RegisterNoProtocolService(s)

	// start server
	etrpc.RunServer(s)

	wg.Wait()
}
