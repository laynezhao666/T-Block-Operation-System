package strategy

import (
	"context"

	pb "trpcprotocol/alarm-server"
)

// IStrategyLogicApi 告警逻辑接口
type IStrategyLogicApi interface {
	GetAlarmName(ctx context.Context, req *pb.ReqAlarmName) (*pb.RspAlarmName, error)
	GetStrategyList(ctx context.Context, req *pb.ReqStrategyList) (*pb.RspStrategyList, error)
	GetStrategyInstance(ctx context.Context, req *pb.ReqStrategyInstance) (*pb.RspStrategyInstance, error)
	GetValidate(ctx context.Context, req *pb.ReqValidateList) (*pb.RspValidateList, error)
	GetVirtualPoint(ctx context.Context, req *pb.ReqGetVirtualPoint) (*pb.RspGetVirtualPoint, error)
}

// NewStrategyLogicApi 新建告警逻辑接口
func NewStrategyLogicApi() IStrategyLogicApi {
	return &strategyLogicImpl{}
}

type strategyLogicImpl struct {
}
