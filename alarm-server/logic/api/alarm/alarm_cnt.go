// Package alarm impl
package alarm

import (
	"context"
	"fmt"
	"time"

	"etrpc-go/log"

	"trpc.group/trpc-go/trpc-go"

	"alarm-server/repo/dao/alarm"

	pb "trpcprotocol/alarm-server"
)

func getActiveAlarmCnt(req *pb.ReqAlarmCnt) (int32, error) {
	begin, end := req.GetBegin(), req.GetEnd()
	con := &alarm.ActiveCntFilter{
		MozuId: req.GetMozuId(),
		Begin:  begin,
		End:    end,
		Level:  req.GetLevel(),
	}
	switch req.GetEventStatus() {
	case 1:
		con.EventStatus = []int64{0}
	case 2:
		con.EventStatus = []int64{1}
	case 3:
		con.EventStatus = []int64{2}
	default:
		con.EventStatus = []int64{}
	}
	switch req.GetAlarmType() {
	case 1:
		con.Status = []int64{0}
	case 2:
		con.Status = []int64{1}
	default:
		con.Status = []int64{}
	}
	// 活动告警，begin和end相同，查询截止至end时的活动告警数量
	// 活动告警，begin和end不同，查询begin到end时间区间内的活动告警，每隔interval统计一次
	ret, err := alarm.NewAlarmDao().GetActiveAlarmCnt(trpc.BackgroundContext(), con)
	if err != nil {
		log.Errorf("GetActiveAlarmCnt err:%s", err.Error())
		return 0, err
	}
	return ret, nil
}

func getHistoryAlarmCnt(req *pb.ReqAlarmCnt) (int32, error) {
	begin, end := req.GetBegin(), req.GetEnd()
	con := &alarm.HistoryCntFilter{
		MozuId: req.GetMozuId(),
		Begin:  begin,
		End:    end,
		Level:  req.GetLevel(),
	}
	ret, err := alarm.NewAlarmDao().GetHistoryAlarmCnt(trpc.BackgroundContext(), con)
	if err != nil {
		log.Errorf("GetHistoryAlarmCnt err:%s", err.Error())
		return 0, err
	}
	return ret, nil
}

// GetAlarmCnt 查询告警数量
func (a *alarmLogicImpl) GetAlarmCnt(ctx context.Context, req *pb.ReqAlarmCnt) (*pb.RspAlarmCnt, error) {
	alarmType := req.GetAlarmType()
	rsp := &pb.RspAlarmCnt{
		Begin: time.Unix(req.GetBegin(), 0).Format(time.DateTime),
		End:   time.Unix(req.GetEnd(), 0).Format(time.DateTime),
	}
	if alarmType == 3 {
		ret, err := getHistoryAlarmCnt(req)
		if err != nil {
			return nil, fmt.Errorf("获取历史告警数量失败: %s", err.Error())
		}
		rsp.Count = ret
	} else if alarmType == 2 || alarmType == 1 {
		ret, err := getActiveAlarmCnt(req)
		if err != nil {
			return nil, fmt.Errorf("获取活动告警数量失败: %s", err.Error())
		}
		rsp.Count = ret
	} else {
		hRet, hErr := getHistoryAlarmCnt(req)
		aRet, aErr := getActiveAlarmCnt(req)
		if hErr != nil || aErr != nil {
			return nil, fmt.Errorf("获取告警数量失败: %s-%s", hErr.Error(), aErr.Error())
		}
		rsp.Count = hRet + aRet
	}
	return rsp, nil
}

// GetAlarmCntTrend 查询24小时内告警数量趋势
func (a *alarmLogicImpl) GetAlarmCntTrend(ctx context.Context, req *pb.ReqAlarmCntTrend) (*pb.RspAlarmCntTrend, error) {
	retList, err := alarm.NewAlarmDao().GetAlarmCntTrend(ctx, int32(req.GetMozuId()))
	if err != nil {
		return nil, err
	}
	return &pb.RspAlarmCntTrend{
		List: retList,
	}, nil
}
