package main

import (
	"cmdb/service"
	"etrpc-go"
	"trpcprotocol/cmdb"
)

func main() {
	// 创建服务
	s := etrpc.NewServer()

	// 注册服务接口
	cmdb.RegisterConfigBuildService(s, service.NewConfigBuildService())
	cmdb.RegisterConfigQueryService(s, service.NewConfigQueryService())

	// 启动服务
	etrpc.RunServer(s)
}
