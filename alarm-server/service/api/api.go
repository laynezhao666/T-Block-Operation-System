// Package api 提供api接口
package api

import (
	"context"
	"fmt"

	"etrpc-go/config"

	"google.golang.org/protobuf/types/known/emptypb"

	"alarm-server/logic/api/alarm"
	"alarm-server/logic/api/strategy"

	pb "trpcprotocol/alarm-server"
)

// IServiceAPi 接口定义
type IServiceAPi interface {
	GetAlarmCnt(ctx context.Context, req *pb.ReqAlarmCnt) (*pb.RspAlarmCnt, error)
	GetAlarmCntTrend(ctx context.Context, req *pb.ReqAlarmCntTrend) (*pb.RspAlarmCntTrend, error)
	GetAlarmList(ctx context.Context, req *pb.ReqAlarmList) (*pb.RspAlarmList, error)
	GetAlarmName(ctx context.Context, req *pb.ReqAlarmName) (*pb.RspAlarmName, error)
	GetStrategy(ctx context.Context, req *pb.ReqStrategyList) (*pb.RspStrategyList, error)
	GetStrategyInstance(ctx context.Context, req *pb.ReqStrategyInstance) (*pb.RspStrategyInstance, error)
	GetValidate(ctx context.Context, req *pb.ReqValidateList) (*pb.RspValidateList, error)
	UpdateAlarmStatus(ctx context.Context, req *pb.ReqUpdateAlarmStatus) (*emptypb.Empty, error)
	GetVirtualPoint(ctx context.Context, req *pb.ReqGetVirtualPoint) (*pb.RspGetVirtualPoint, error)
	DelHistoryAlarm(ctx context.Context, req *pb.ReqDelHistoryAlarm) (*emptypb.Empty, error)
	AlarmDiagnose(ctx context.Context, req *pb.ReqAlarmDiagnose) (*pb.RspAlarmDiagnose, error)
}

// NewServiceApi 构造函数
func NewServiceApi() IServiceAPi {
	return &serviceImpl{}
}

// serviceImpl 告警数据库基本查询接口
type serviceImpl struct{}

// GetAlarmCnt 查询告警数量
func (s *serviceImpl) GetAlarmCnt(ctx context.Context, req *pb.ReqAlarmCnt) (*pb.RspAlarmCnt, error) {
	if req.MozuId == 0 {
		return nil, fmt.Errorf("查询告警数量, 无效请求, 模组Id: %+d", req.MozuId)
	}
	return alarm.NewAlarmLogicApi().GetAlarmCnt(ctx, req)
}

func (s *serviceImpl) GetAlarmCntTrend(ctx context.Context, req *pb.ReqAlarmCntTrend) (*pb.RspAlarmCntTrend, error) {
	if req.MozuId == 0 {
		return nil, fmt.Errorf("查询24小时内告警数量趋势, 无效请求, 模组Id: %+d", req.MozuId)
	}
	return alarm.NewAlarmLogicApi().GetAlarmCntTrend(ctx, req)
}

// GetAlarmList 查询告警列表
func (s *serviceImpl) GetAlarmList(ctx context.Context, req *pb.ReqAlarmList) (*pb.RspAlarmList, error) {
	if req.MozuId == 0 {
		return nil, fmt.Errorf("查询告警列表, 无效请求, 模组Id: %+d", req.MozuId)
	}
	return alarm.NewAlarmLogicApi().GetAlarmList(ctx, req)
}

// GetAlarmName 查询告警名称
func (s *serviceImpl) GetAlarmName(ctx context.Context, req *pb.ReqAlarmName) (*pb.RspAlarmName, error) {
	return strategy.NewStrategyLogicApi().GetAlarmName(ctx, req)
}

// GetStrategy 查询策略列表
func (s *serviceImpl) GetStrategy(ctx context.Context, req *pb.ReqStrategyList) (*pb.RspStrategyList, error) {
	if req.MozuId == 0 {
		return nil, fmt.Errorf("查询策略列表, 无效请求, 模组Id: %+d", req.MozuId)
	}
	return strategy.NewStrategyLogicApi().GetStrategyList(ctx, req)
}

// GetStrategyInstance 查询策略实例列表
func (s *serviceImpl) GetStrategyInstance(ctx context.Context, req *pb.ReqStrategyInstance) (*pb.RspStrategyInstance, error) {
	if req.MozuId == 0 {
		return nil, fmt.Errorf("查询策略实例列表, 无效请求, 模组Id: %+d", req.MozuId)
	}
	return strategy.NewStrategyLogicApi().GetStrategyInstance(ctx, req)
}

// GetValidate 查询策略生效信息
func (s *serviceImpl) GetValidate(ctx context.Context, req *pb.ReqValidateList) (*pb.RspValidateList, error) {
	if req.MozuId == 0 {
		return nil, fmt.Errorf("查询策略生效信息, 无效请求, 模组Id: %+d", req.MozuId)
	}
	return strategy.NewStrategyLogicApi().GetValidate(ctx, req)
}

// UpdateAlarmStatus 更新告警状态
func (s *serviceImpl) UpdateAlarmStatus(ctx context.Context, req *pb.ReqUpdateAlarmStatus) (*emptypb.Empty, error) {
	if req.MozuId == 0 {
		return nil, fmt.Errorf("更新告警状态, 无效请求, 模组Id: %+d", req.MozuId)
	}
	return alarm.NewAlarmLogicApi().UpdateAlarmStatus(ctx, req)
}

// GetVirtualPoint 查询虚拟测点
func (s *serviceImpl) GetVirtualPoint(ctx context.Context, req *pb.ReqGetVirtualPoint) (*pb.RspGetVirtualPoint, error) {
	if req.MozuId == 0 {
		return nil, fmt.Errorf("查询虚拟测点, 无效请求, 模组Id: %+d", req.MozuId)
	}
	return strategy.NewStrategyLogicApi().GetVirtualPoint(ctx, req)
}

// DelHistoryAlarm 删除历史告警
func (s *serviceImpl) DelHistoryAlarm(ctx context.Context, req *pb.ReqDelHistoryAlarm) (*emptypb.Empty, error) {
	// token校验，防止误删
	configToken, ok := config.GetString("db_admin.del_token")
	if !ok || req.GetToken() != configToken {
		return nil, fmt.Errorf("删除失败: token校验失败")
	}
	return alarm.NewAlarmLogicApi().DelHistoryAlarm(ctx, req)
}

// AlarmDiagnose 告警诊断
func (s *serviceImpl) AlarmDiagnose(ctx context.Context, req *pb.ReqAlarmDiagnose) (*pb.RspAlarmDiagnose, error) {
	if req.MozuId == 0 {
		return nil, fmt.Errorf("告警诊断, 无效请求, 模组Id: %+d", req.MozuId)
	}
	return alarm.NewAlarmLogicApi().AlarmDiagnose(ctx, req)
}
