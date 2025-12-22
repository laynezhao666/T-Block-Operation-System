package main

import (
	"collector/logic/bus/collectors_config"
	"collector/logic/bus/data/collect_point"
	"collector/logic/bus/data/tbos_point"
	"collector/repo"
	"collector/repo/report"
	"collector/service"
	"etrpc-go"
	pb "trpcprotocol/collector"
	monitorPb "trpcprotocol/tboxmonitor"
)

func main() {
	s := etrpc.NewServer()
	collect_point.Init()
	tbos_point.Init()
	if err := repo.Init(); err != nil {
		panic(err)
	}
	collectors_config.Init()
	report.Init()

	pb.RegisterConfigBusService(s, &service.ConfigBusServiceImpl{})
	pb.RegisterDataBusService(s, &service.DatabusServiceImpl{})
	pb.RegisterCollectPointForwardService(s, &service.CollectPointForwardServiceImpl{})
	pb.RegisterExternalPlatformService(s, &service.ExternalPlatformServiceImpl{})
	monitorPb.RegisterMonitorService(s, &service.ControlBusServiceImpl{})

	// baTest(s)
	etrpc.RunServer(s)
}
