package alarm

import (
	"context"

	pb "trpcprotocol/alarm-server"

	"google.golang.org/protobuf/types/known/emptypb"
)

// IAlarmLogicApi 告警逻辑接口
type IAlarmLogicApi interface {
	GetAlarmCnt(ctx context.Context, req *pb.ReqAlarmCnt) (*pb.RspAlarmCnt, error)
	GetAlarmCntTrend(ctx context.Context, req *pb.ReqAlarmCntTrend) (*pb.RspAlarmCntTrend, error)
	GetAlarmList(ctx context.Context, req *pb.ReqAlarmList) (*pb.RspAlarmList, error)
	UpdateAlarmStatus(ctx context.Context, req *pb.ReqUpdateAlarmStatus) (*emptypb.Empty, error)
	DelHistoryAlarm(ctx context.Context, req *pb.ReqDelHistoryAlarm) (*emptypb.Empty, error)
	AlarmDiagnose(ctx context.Context, req *pb.ReqAlarmDiagnose) (*pb.RspAlarmDiagnose, error)
}

// NewAlarmLogicApi 新建告警逻辑接口
func NewAlarmLogicApi() IAlarmLogicApi {
	return &alarmLogicImpl{}
}

type alarmLogicImpl struct {
}
