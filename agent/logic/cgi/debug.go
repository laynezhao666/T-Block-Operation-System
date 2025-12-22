package cgi

import (
	"context"

	"agent/logic/debug"

	pb "trpcprotocol/agent"
)

// DebugHandle debug
func DebugHandle(ctx context.Context, req *pb.ReqDebug) (*pb.RspDebug, error) {
	t := req.EnableTime
	debug.SetEnable(int(t))
	return &pb.RspDebug{
		Msg: "debug is enabled",
	}, nil
}
