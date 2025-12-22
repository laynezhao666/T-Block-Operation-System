package cgi

import (
	"context"

	pb "trpcprotocol/agent"
)

// StartupProbeHandle 服务启动探测
func StartupProbeHandle(ctx context.Context) (*pb.RspStartupProbe, error) {
	return &pb.RspStartupProbe{
		Msg: "ok",
	}, nil
}
