// Package consumer 同步告警消息
package consumer

import (
	"context"
	"encoding/json"

	"etrpc-go/log"

	"github.com/IBM/sarama"

	"cgi/logic/alarm/wslogic"
	cmodel "common/entity/model"
)

// BatchHandleMessage 批量处理消息
func BatchHandleMessage(ctx context.Context, msgs []*sarama.ConsumerMessage) error {
	msgsList := make([][]cmodel.AlarmActive, 0)
	for _, msg := range msgs {
		activeList := []cmodel.AlarmActive{}
		err := json.Unmarshal(msg.Value, &activeList)
		if err != nil {
			log.Errorf("Unmarshal kafka message failed, err:%s", err.Error())
			continue
		}
		if len(activeList) > 0 {
			msgsList = append(msgsList, activeList)
		}
	}
	mozuIdSet := map[int32]struct{}{}
	for _, activeList := range msgsList {
		for _, active := range activeList {
			if _, ok := mozuIdSet[int32(active.MozuId)]; !ok {
				wslogic.GetAlarmWSImpl().AddMozuPushTask(int32(active.MozuId))
				mozuIdSet[int32(active.MozuId)] = struct{}{}
			}
		}
	}
	return nil
}
