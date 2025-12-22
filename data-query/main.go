package main

import (
	"data-query/repo/read"
	"data-query/service"
	"etrpc-go"
	"trpcprotocol/data-query"
)

func main() {
	s := etrpc.NewServer()

	read.Init()
	// register your service here, build your pb definition, then call pb.RegisterXXXX
	// pb.RegisterDemoService(s, &service.DemoService{})
	data_query.RegisterDataService(s, &service.DataServiceImpl{})

	etrpc.RunServer(s)
}
