package service

import (
	"context"
	"fmt"
	"agent/entity/errcode"
	"agent/logic/cgi"
	httpDt "agent/logic/distribution/distributor/http"

	pb "trpcprotocol/agent"

	"google.golang.org/protobuf/types/known/emptypb"
	"trpc.group/trpc-go/trpc-go/errs"
)

const (
	TypeKafka string = "kafka"
	TypeHttp  string = "http"
)

// RealTimeDataServiceImpl /north 接口实现
type RealTimeDataServiceImpl struct{}

// SetMessagePushParams /north 设置消息推送参数
func (r *RealTimeDataServiceImpl) SetMessagePushParams(ctx context.Context, req *pb.SetMessagePushParamsReq) (*emptypb.Empty, error) {
	switch req.GetType() {
	case TypeKafka:
		return &emptypb.Empty{}, errs.New(errcode.ErrCgiParamInvalid, "Kafka type not supported yet")
	case TypeHttp:
		err := httpDt.HttpDistributor().AddClient(req.GetType(), req.GetName(), req.GetTarget(), req.GetClientId())
		if err != nil {
			return &emptypb.Empty{}, errs.New(errcode.ErrCgiParamInvalid, err.Error())
		}
		return &emptypb.Empty{}, nil
	default:
		return &emptypb.Empty{}, errs.New(errcode.ErrCgiParamInvalid, fmt.Sprintf("unknown type: [%v]", req.GetType()))
	}
}

// GetMessagePushParams /north/get
func (r *RealTimeDataServiceImpl) GetMessagePushParams(ctx context.Context, req *emptypb.Empty) (*pb.GetMessagePushParamsRsp, error) {
	configs := httpDt.HttpDistributor().GetAllClientsConfig()
	params := make([]*pb.GetMessagePushParamsRsp_MessagePushParams, 0, len(configs))
	for _, c := range configs {
		params = append(params, &pb.GetMessagePushParamsRsp_MessagePushParams{
			Type:     c.Type,
			Target:   c.Target,
			ClientId: c.ClientId,
			Name:     c.Name,
		})
	}
	return &pb.GetMessagePushParamsRsp{
		Params: params,
	}, nil
}

// OnlineStrategyPush /north/online_strategy_push
func (r *RealTimeDataServiceImpl) OnlineStrategyPush(ctx context.Context, req *pb.OnlineStrategyPushReq) (*pb.OnlineStrategyPushRsp, error) {
	if err := cgi.OnlinePushPointHandle(req); err != nil {
		return &pb.OnlineStrategyPushRsp{
			Status: false,
		}, nil
	}
	return &pb.OnlineStrategyPushRsp{
		Status: true,
	}, nil
}
