package service

import (
	"context"

	"cgi/logic/alarm/api"

	pb "trpcprotocol/cgi"
)

type alarmService struct {
	alarmApi api.IAlarmApi
}

// NewAlarmService 创建一个Cmdb接口服务对象
func NewAlarmService() pb.AlarmService {
	return &alarmService{
		alarmApi: api.NewAlarmApi(),
	}
}

func (a *alarmService) GetAlarmCnt(ctx context.Context, req *pb.ReqAlarmCnt) (*pb.RspAlarmCnt, error) {
	return a.alarmApi.GetAlarmCnt(ctx, req)
}

func (a *alarmService) GetAlarmCntTrend(ctx context.Context, req *pb.ReqAlarmCntTrend) (*pb.RspAlarmCntTrend, error) {
	return a.alarmApi.GetAlarmCntTrend(ctx, req)
}

func (a *alarmService) GetAlarmName(ctx context.Context, req *pb.ReqAlarmName) (*pb.RspAlarmName, error) {
	return a.alarmApi.GetAlarmName(ctx, req)
}

func (a *alarmService) GetAlarmList(ctx context.Context, req *pb.ReqAlarmList) (*pb.RspAlarmList, error) {
	return a.alarmApi.GetAlarmList(ctx, req)
}

func (a *alarmService) GetStrategy(ctx context.Context, req *pb.ReqStrategyList) (*pb.RspStrategyList, error) {
	return a.alarmApi.GetStrategy(ctx, req)
}

func (a *alarmService) GetStrategyInstance(ctx context.Context, req *pb.ReqStrategyInstance) (*pb.RspStrategyInstance, error) {
	return a.alarmApi.GetStrategyInstance(ctx, req)
}

func (a *alarmService) GetValidate(ctx context.Context, req *pb.ReqValidateList) (*pb.RspValidateList, error) {
	return a.alarmApi.GetValidate(ctx, req)
}

func (a *alarmService) GetVirtualPoint(ctx context.Context, req *pb.ReqGetVirtualPoint) (*pb.RspGetVirtualPoint, error) {
	return a.alarmApi.GetVirtualPoint(ctx, req)
}
