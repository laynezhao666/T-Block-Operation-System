package service

import (
	"collector/entity/errcode"
	"collector/logic/bus/data/collect_point"
	"context"

	pb "trpcprotocol/collector"

	"google.golang.org/protobuf/types/known/emptypb"
	"trpc.group/trpc-go/trpc-go"
	"trpc.group/trpc-go/trpc-go/errs"
)

// CollectPointForwardServiceImpl 数据总线
type CollectPointForwardServiceImpl struct{}

// Forward 过渡办法，转发采集测点到原有动环
func (c *CollectPointForwardServiceImpl) Forward(ctx context.Context, req *pb.ReqSend) (*emptypb.Empty, error) {
	err := trpc.Go(ctx, defaultSendDataTimeout, func(ctx context.Context) {
		collect_point.ForwardHandle(ctx, req)
	})
	if err != nil {
		return nil, errs.New(errcode.ErrSendFail, err.Error())
	}
	return &emptypb.Empty{}, nil
}
