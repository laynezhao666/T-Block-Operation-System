// Package etrpc 框架基础配置处理
package etrpc

import (
	"etrpc-go/config"
	"trpc.group/trpc-go/trpc-go"
)

// Etrpc Etrpc配置项详情
type Etrpc struct {
	ServiceName     string `yaml:"service_name"`      // etRPC服务名称
	ServicePort     int32  `yaml:"service_port"`      // http协议端口,默认8080
	TrpcServicePort int32  `yaml:"trpc_service_port"` // trpc协议端口,默认8081
}

var (
	CfgTrpc  = &trpc.Config{} // Trpc配置
	CfgEtrpc = &Etrpc{}       // Etrpc配置
)

func init() {
	config.RegisterConfig("trpc.framework", CfgTrpc, false)
	config.RegisterConfigWithPrefix("etrpc.framework", "etrpc", CfgEtrpc, false)
}
