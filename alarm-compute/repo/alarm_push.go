package repo

import (
	"context"

	pb "trpcprotocol/alarm-compute"
	managePb "trpcprotocol/alarm-manage"
)

// IAlarmManageApi 告警相关接口
type IAlarmManageApi interface {
	PushAlarmByApi(ctx context.Context, req *pb.FireAlertMsg) error
}

type alarmManageApi struct {
	manageClientProxy managePb.ManageClientProxy
}

// NewAlarmManageApi 创建Alarm相关逻辑接口实现类
func NewAlarmManageApi() IAlarmManageApi {
	return &alarmManageApi{
		manageClientProxy: managePb.NewManageClientProxy(),
	}
}

// PushAlarmByApi 当发送kafka失败时，调用此接口发送
func (a *alarmManageApi) PushAlarmByApi(ctx context.Context, fireAlert *pb.FireAlertMsg) error {
	req := &managePb.AlarmMsgPb{
		StartAt:       fireAlert.StartAt,
		EndAt:         fireAlert.EndAt,
		Rid:           fireAlert.Rid,
		Gid:           fireAlert.Gid,
		DeviceNumber:  fireAlert.DeviceNumber,
		Level:         fireAlert.Level,
		AlarmName:     fireAlert.AlarmName,
		Content:       fireAlert.Content,
		MozuId:        fireAlert.MozuId,
		Env:           fireAlert.Env,
		AnalyzeResult: fireAlert.AnalyzeResult,
	}
	_, err := a.manageClientProxy.PushAlarm(ctx, req)
	if err != nil {
		return err
	}
	return nil
}
