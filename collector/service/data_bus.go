package service

import (
	"context"
	"time"

	"collector/entity/errcode"
	"collector/logic/bus/data/tbos_point"

	pb "trpcprotocol/collector"

	"google.golang.org/protobuf/types/known/emptypb"
	"trpc.group/trpc-go/trpc-go"
	"trpc.group/trpc-go/trpc-go/errs"
)

const (
	defaultSendDataTimeout time.Duration = 10 * time.Second
)

// DatabusServiceImpl 数据总线
type DatabusServiceImpl struct{}

// Send 接口实现，接收数据并上报
func (c *DatabusServiceImpl) Send(ctx context.Context, req *pb.ReqSend) (*emptypb.Empty, error) {
	err := trpc.Go(ctx, defaultSendDataTimeout, func(ctx context.Context) {
		tbos_point.SendHandle(ctx, req)
	})
	if err != nil {
		return nil, errs.New(errcode.ErrSendFail, err.Error())
	}
	return &emptypb.Empty{}, nil
}
