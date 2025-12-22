package service

import (
	"context"
	"fmt"

	pb "trpcprotocol/alarm-compute"

	"google.golang.org/protobuf/types/known/emptypb"

	"alarm-compute/logic/diagnose"
	"alarm-compute/logic/heartbeat"
	"alarm-compute/logic/strategy"
)

// IAlarmService IAlarmService
type IAlarmService interface {
	RecvTask(ctx context.Context, req *pb.ReqStrategyRecv) (*emptypb.Empty, error)
	ExpCompute(ctx context.Context, req *pb.ReqExpCompute) (*pb.RspExpCompute, error)
}

// NewAlarmServiceImpl NewAlarmServiceImpl
func NewAlarmServiceImpl() IAlarmService {
	return &alarmServiceImpl{}
}

// alarmServiceImpl AlarmService接口实现类
type alarmServiceImpl struct {
}

// RecvTask receive alarm strategy task from scheduler
func (*alarmServiceImpl) RecvTask(ctx context.Context, req *pb.ReqStrategyRecv) (*emptypb.Empty, error) {
	successAdd := strategy.GetStrategyHandler().AddStrategyReq(req)
	if !successAdd {
		return nil, fmt.Errorf("add strategy task fail")
	}
	heartbeat.GetHeartAgent().SetTaskVerMask(ctx, req.RecvTimestamp, req.RecvVersion)
	return &emptypb.Empty{}, nil
}

// ExpCompute 表达式计算，用于告警回放/诊断
func (*alarmServiceImpl) ExpCompute(ctx context.Context, req *pb.ReqExpCompute) (*pb.RspExpCompute, error) {
	if req.Interval == 0 || req.BeginTime > req.EndTime || req.BeginTime == 0 || req.EndTime == 0 {
		return nil, fmt.Errorf("invalid input, req: %+v", req)
	}
	return diagnose.NewDiagnoseSvc().ExpCompute(ctx, req)
}
