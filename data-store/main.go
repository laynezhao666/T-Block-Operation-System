package main

import (
	"common/entity/consts"
	"data-store/repo/store"
	"data-store/service"
	"etrpc-go"

	"trpc.group/trpc-go/trpc-database/kafka"
)

func main() {
	// 创建 etrpc 服务器
	s := etrpc.NewServer()

	// 初始化存储插件
	store.Init()

	// 注册 kafka 消费者服务
	if majorKafkaSvr := s.Service(consts.TbosMajorKafkaName); majorKafkaSvr != nil {
		kafka.RegisterBatchHandlerService(majorKafkaSvr, service.MajorKafkaHandle)
	}
	if backUpKafkaSvr := s.Service(consts.TbosBackupKafkaName); backUpKafkaSvr != nil {
		kafka.RegisterBatchHandlerService(backUpKafkaSvr, service.BackupKafkaHandle)
	}

	// 运行 etrpc 服务器
	etrpc.RunServer(s)

	// 服务停止时关闭存储插件接口
	store.Close()
}
