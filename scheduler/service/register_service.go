package service

import (
	"context"
	"fmt"
	"google.golang.org/protobuf/types/known/emptypb"
	"math"
	"scheduler/logic/register"
	"trpc.group/trpc-go/trpc-go"
	"trpcprotocol/scheduler"
)

// NewRegisterService 创建一个Worker注册服务
func NewRegisterService() scheduler.RegisterService {
	return &registerServiceImpl{
		registerApi: register.NewRegisterApi(),
	}
}

type registerServiceImpl struct {
	registerApi register.IRegisterApi
}

func (r registerServiceImpl) Heartbeat(ctx context.Context, req *scheduler.WorkerInfo) (*emptypb.Empty, error) {
	if req.WorkerType == scheduler.WorkerInfo_INVALID {
		return nil, fmt.Errorf("bad worker_type:[%d,%s]", req.WorkerType.Number(), req.WorkerType.String())
	}
	// IP默认调用方IP
	if req.Ip == "" {
		msg := trpc.Message(ctx)
		req.Ip = msg.RemoteAddr().String()
	}
	// 端口tRPC协议默认8081,其他协议默认8080
	if req.Port == 0 {
		if req.WorkerProtocol == scheduler.WorkerInfo_TRPC {
			req.Port = 8081
		}
		req.Port = 8080
	}
	// 最大处理能力没传递则为无限,用math.MaxInt32代替
	if req.MaxProcessCap == 0 {
		req.MaxProcessCap = math.MaxInt32
	}
	if err := r.registerApi.Heartbeat(ctx, req); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}
