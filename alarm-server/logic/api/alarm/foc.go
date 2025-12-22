package alarm

import (
	"context"
	"fmt"

	"google.golang.org/protobuf/types/known/emptypb"

	"alarm-server/repo/dao/alarm"

	pb "trpcprotocol/alarm-server"
)

// UpdateAlarmStatus 更新告警状态
// @param ctx context.Context
// @param req *pb.ReqUpdateAlarmStatus
// req.OpType 0:挂起/解除挂起 1: 转单/结单 2：关闭告警
func (a *alarmLogicImpl) UpdateAlarmStatus(ctx context.Context, req *pb.ReqUpdateAlarmStatus) (*emptypb.Empty, error) {
	var err error
	if len(req.AlarmIds) > 1000 || len(req.AlarmIds) == 0 {
		return nil, fmt.Errorf("please limit the length of alarm_ids to (0, 1000]")
	}
	switch req.OpType {
	case pb.ReqUpdateAlarmStatus_HANGOP:
		// 修改alarm status
		con := &alarm.AlarmStatusCon{
			MozuId:       req.MozuId,
			AlarmIds:     req.AlarmIds,
			UserId:       req.UserId,
			AlarmStatus:  req.AlarmStatus,
			HangupReason: req.HangupReason,
			UpdateTime:   req.UpdateTime,
		}
		// 挂起时 挂起原因和user_id不能为空
		if con.AlarmStatus == 1 && (len(con.HangupReason) == 0 || con.UserId <= 0) {
			return nil, fmt.Errorf("invalid hangup reason or user_id")
		}
		err = alarm.NewAlarmDao().UpdateAlarmStatus(ctx, int64(req.OpType), con)
	case pb.ReqUpdateAlarmStatus_EVENTOP:
		// 修改alarm event status
		con := &alarm.AlarmStatusCon{
			MozuId:      req.MozuId,
			AlarmIds:    req.AlarmIds,
			EventStatus: req.EventStatus,
			UpdateTime:  req.UpdateTime,
		}
		err = alarm.NewAlarmDao().UpdateAlarmStatus(ctx, int64(req.OpType), con)
	case pb.ReqUpdateAlarmStatus_CLOSEOP:
		// close alarm
		con := &alarm.CloseStatusCon{
			MozuId:      req.MozuId,
			AlarmIds:    req.AlarmIds,
			UserId:      req.UserId,
			CloseReason: req.CloseReason,
		}
		if len(con.CloseReason) == 0 || con.UserId <= 0 {
			return nil, fmt.Errorf("invalid close reason or user_id")
		}
		err = alarm.NewAlarmDao().CloseAlarms(ctx, con)
	default:
		return nil, fmt.Errorf("invalid op_type: %d", req.OpType)
	}
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}
