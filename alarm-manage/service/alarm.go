package service

import (
	"context"
	"fmt"

	pb "trpcprotocol/alarm-manage"

	"etrpc-go/log"

	"google.golang.org/protobuf/types/known/emptypb"
	"trpc.group/trpc-go/trpc-go"

	"alarm-manage/logic/manager"
	"alarm-manage/logic/notification"
	"alarm-manage/repo/db"
)

// PushAlarm 推送告警
// 当compute发送kafka失败时，调用此接口发送告警
func (o *ManageService) PushAlarm(ctx context.Context, req *pb.AlarmMsgPb) (*emptypb.Empty, error) {
	if req.EndAt > 0 {
		// 恢复告警
		manager.GetGlobalManager().AddRestoreToCh(req)
	} else {
		manager.GetGlobalManager().AddAlertToCh(req)
	}
	return &emptypb.Empty{}, nil
}

// ResendAlarm 重新推送全量告警
// 服务重新启动/发布时，需要调用该接口
func (o *ManageService) ResendAlarm(ctx context.Context, req *pb.ReqResendAlarm) (*emptypb.Empty, error) {
	var alarmStatus int
	alarmStatus = int(req.RescendType)
	if alarmStatus > 1 {
		alarmStatus = 0
	}
	alarms, err := db.GetAlarmDBImpl().GetTotalActiveList(trpc.BackgroundContext(), req.MozuId, alarmStatus)
	if err != nil {
		log.Errorf("GetTotalActiveList failed, err: %s", err.Error())
		return nil, fmt.Errorf("GetTotalActiveList failed, err: %s", err.Error())
	}
	err = notification.GetSpeaker().BatchReportAlert(alarms)
	if err != nil {
		log.Errorf("BatchReportAlert failed, err: %s", err.Error())
		return nil, fmt.Errorf("发送告警到云端kafka失败, err: %s", err.Error())
	}
	return &emptypb.Empty{}, nil
}
