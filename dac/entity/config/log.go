// Package config 提供门禁服务的全局配置和日志实例。
package config

import (
	"trpc.group/trpc-go/trpc-go/log"
)

// Log 全局日志实例，使用tRPC默认日志器
var (
	Log = log.GetDefaultLogger()
)
