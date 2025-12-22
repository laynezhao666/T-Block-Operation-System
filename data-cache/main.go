package main

import (
	"data-cache/entity/consts"
	"data-cache/logic/kafkasvr"
	"data-cache/repo/cache"
	"data-cache/service"
	"etrpc-go"
	"fmt"
	"sync"
	"time"
	"trpcprotocol/data-cache"

	"trpc.group/trpc-go/trpc-database/kafka"
	"trpc.group/trpc-go/trpc-go/log"
	"trpc.group/trpc-go/trpc-go/server"
)

func main() {
	// 创建 etrpc 服务器
	s := etrpc.NewServer()

	// 初始化缓存服务
	cache.Setup()

	// 初始化并注册数据服务
	data_cache.RegisterPointService(s, service.NewPointService())

	// 注册kafka消费服务，启动kafka测点消费者服务,并等待消费到当前时间
	wg := &sync.WaitGroup{}
	if majorKafkaSvr := s.Service(consts.TbosMajorKafkaName); majorKafkaSvr != nil {
		kafka.RegisterBatchHandlerService(majorKafkaSvr, service.MajorKafkaHandle)
		startKafka(s, consts.TbosMajorKafkaName, wg)
	}
	if backupKafkaSvr := s.Service(consts.TbosBackupKafkaName); backupKafkaSvr != nil {
		kafka.RegisterBatchHandlerService(backupKafkaSvr, service.BackupKafkaHandle)
		startKafka(s, consts.TbosBackupKafkaName, wg)
	}
	wg.Wait()

	// 运行 etrpc 服务器
	etrpc.RunServer(s)

}

func startKafka(s *server.Server, serviceName string, wg *sync.WaitGroup) {
	go func() {
		err := s.Service(serviceName).Serve()
		if err != nil {
			panic(fmt.Errorf("start kafka service [%s] fail, err: [%v]", serviceName, err))
		}
	}()
	wg.Add(1)
	go func() {
		begin := time.Now()
		if err := kafkasvr.WaitReady(serviceName); err != nil {
			panic(fmt.Errorf("wait kafka ready fail, err: %v", err))
		}
		log.Infof("kafka [%s] finish consume history data, cost: %.0fs", serviceName, time.Since(begin).Seconds())
		wg.Done()
	}()
}
