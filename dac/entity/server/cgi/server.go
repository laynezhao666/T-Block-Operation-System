// Package cgi 提供门禁系统的HTTP CGI服务注册和路由配置。
package cgi

import (
	"trpc.group/trpc-go/trpc-go/http"
	"trpc.group/trpc-go/trpc-go/server"
)

// Register 注册CGI服务到tRPC HTTP服务，禁用自动读取Body以支持gin框架
func Register(s server.Service) {
	http.DefaultServerCodec.AutoReadBody = false
	http.RegisterNoProtocolServiceMux(s, getHandler())
}
