// Package etrpc is the Go implementation of tRPC, which is designed to be high-performance,
// everything-pluggable and easy for testing.
package etrpc

import (
	"etrpc-go/config/loader"
	"etrpc-go/healthcheck"
	"go.uber.org/automaxprocs/maxprocs"
	"trpc.group/trpc-go/trpc-go"
	"trpc.group/trpc-go/trpc-go/log"
	"trpc.group/trpc-go/trpc-go/server"

	// trpc plugin
	_ "trpc.group/trpc-go/trpc-metrics-runtime"

	// trpc filter
	_ "etrpc-go/filter/rsp"
	_ "trpc.group/trpc-go/trpc-filter/recovery"
	_ "trpc.group/trpc-go/trpc-filter/validation"
)

// NewServer returns tRPC server
func NewServer(opt ...server.Option) *server.Server {
	// 加载配置
	loader.LoadConfig()

	// 启动前置校验,Trpc的Service列表不允许为空
	if len(CfgTrpc.Server.Service) == 0 {
		panic("start etrpc server failed, no trpc service config found")
	}
	// 服务名称必须配置
	if len(CfgEtrpc.ServiceName) == 0 {
		panic("start etrpc server failed, config `etrpc.service_name` can not be empty")
	}

	s := newServerWithCfg(CfgTrpc, opt...)

	return s
}

func newServerWithCfg(cfg *trpc.Config, opt ...server.Option) *server.Server {
	// 修复配置
	if err := trpc.RepairConfig(cfg); err != nil {
		panic("repair config fail: " + err.Error())
	}

	// set to global config for other plugins' accessing to the config
	trpc.SetGlobalConfig(cfg)

	closePlugins, err := trpc.SetupPlugins(cfg.Plugins)
	if err != nil {
		panic("setup plugin fail: " + err.Error())
	}
	if err := trpc.SetupClients(&cfg.Client); err != nil {
		panic("failed to setup client: " + err.Error())
	}
	// keep backward compatible with Setup.
	//plugin.SetupFinished()

	// set default GOMAXPROCS for docker
	_, _ = maxprocs.Set(maxprocs.Logger(log.Debugf))
	s := trpc.NewServerWithConfig(cfg, opt...)

	// check db client is ok
	if err = healthcheck.CheckClients(cfg); err != nil {
		panic(err)
	}

	s.RegisterOnShutdown(func() {
		if err := closePlugins(); err != nil {
			log.Errorf("failed to close plugins, err: %s", err)
		}
	})
	return s
}

// RunServer 执行服务启动,预留Hook，方便在此做一些事情
func RunServer(s *server.Server) {
	if err := s.Serve(); err != nil {
		log.Fatalf("server start fail, %s", err)
	}
}
