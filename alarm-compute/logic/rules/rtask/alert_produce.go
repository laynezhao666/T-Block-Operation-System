package rtask

import (
	"encoding/json"
	"fmt"

	"etrpc-go/log"
	pb "trpcprotocol/alarm-compute"

	"google.golang.org/protobuf/proto"
	"trpc.group/trpc-go/trpc-go"

	"alarm-compute/conf"
	"alarm-compute/entity/epoint"
	"alarm-compute/logic/lcache"
	"alarm-compute/logic/pointeval"
	"alarm-compute/repo"
	"alarm-compute/utils/modcall"
)

// GeneFireAlert 生成告警信息
func (at *AlarmTask) GeneFireAlert(ts int64, pointValueMap map[string]float64,
	historyPointValueMap epoint.HistoryValueMap) (*pb.FireAlertMsg, error) {
	result := pointeval.AlarmTaskRet{
		PointValueMap:        pointValueMap,
		HistoryPointValueMap: historyPointValueMap,
		PointMap:             at.Exp.PMap,
		StartRunAt:           ts,
		ExpMap:               at.Exp,
	}
	retJson, err := json.Marshal(result)
	if err != nil {
		log.Warnf("序列化告警结果失败, key: %s, reasion:%s", at.RuleTask.GetKey(), err.Error())
		retJson = []byte{}
	}
	a := &pb.FireAlertMsg{
		Rid:           at.RuleTask.Rid,
		Gid:           at.RuleTask.Gid,
		Level:         at.RuleTask.Level,
		AlarmName:     at.RuleTask.AlarmName,
		Content:       at.RuleTask.ContentTemplate,
		MozuId:        int32(at.RuleTask.MozuId),
		AnalyzeResult: string(retJson[:]),
	}
	return a, nil
}

// SendAlert SendAlert
func (rt *RuleTask) SendAlert(startAt int64, fireAlert *pb.FireAlertMsg) error {
	//TODO
	// 更新本地缓存
	// 发送告警
	if fireAlert == nil {
		log.Errorf("告警信息为空, rule key: %s", rt.GetKey())
		return fmt.Errorf("告警信息为空, rule key: %s", rt.GetKey())
	}
	lcache.GetLocalCache().SetActiveAlarmCache(rt.GetKey(), startAt,
		int64(conf.ServerConf.ActiveAlarmCache.CacheKeyTimeDuration))
	fireAlert.StartAt = startAt
	data, err := proto.Marshal(fireAlert)
	if err != nil {
		log.Errorf("发送告警序列化失败:%s", rt.GetKey())
		return err
	}
	key := rt.GetKey()
	// 发送告警
	err = repo.GetCkafka().SendAlertMsg([]byte(key), data)
	if err != nil {
		log.Errorf("发送告警失败:%s", key)
		// 如果kafka发送失败，则调用manage推送接口
		err = repo.NewAlarmManageApi().PushAlarmByApi(trpc.BackgroundContext(), fireAlert)
		if err != nil {
			log.AlarmContextf(trpc.BackgroundContext(),
				"发送告警信息，Kafka和接口均失败。Rid:%d, Gid:%s, start:%d",
				fireAlert.Rid, fireAlert.Gid, fireAlert.StartAt)
			return err
		} else {
			log.Errorf("发送告警信息，Kafka失败，调用接口成功。Rid:%d, Gid:%s, start:%d",
				fireAlert.Rid, fireAlert.Gid, fireAlert.StartAt)
		}
	} else {
		log.Infof("发送告警消息成功,key:%v", string(key))
	}
	modcall.RecordProduceAlertCnt(int(rt.MozuId), 1)
	return nil
}

// SendRestoreAlert SendRestoreAlert
func (rt *RuleTask) SendRestoreAlert(endAt int64, fireAlert *pb.FireAlertMsg) error {
	// TODO
	// 删除本地缓存
	// 发送恢复告警
	if fireAlert == nil {
		log.Errorf("告警恢复信息为空, rule key: %s", rt.GetKey())
		return fmt.Errorf("告警恢复信息为空, rule key: %s", rt.GetKey())
	}
	lcache.GetLocalCache().RestoreAlarmCache(rt.GetKey())
	fireAlert.EndAt = endAt
	data, err := proto.Marshal(fireAlert)
	if err != nil {
		log.Errorf("发送恢复告警序列化失败:%s", rt.GetKey())
		return err
	}
	key := rt.GetKey()
	// 发送告警
	err = repo.GetCkafka().SendAlertMsg([]byte(key), data)
	if err != nil {
		log.Errorf("发送恢复告警消息失败:%s", key)
		return err
	} else {
		log.Infof("发送恢复告警消息成功,key:%v", string(key))
	}
	modcall.RecordProduceAlertCnt(int(rt.MozuId), 1)
	return nil
}
