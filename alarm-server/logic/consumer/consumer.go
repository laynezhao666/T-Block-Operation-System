// Package consumer consumer
package consumer

import (
	"context"

	"etrpc-go/log"
	cpb "trpcprotocol/alarm-compute"

	"github.com/IBM/sarama"
	"google.golang.org/protobuf/proto"

	"alarm-server/entity/model"
	"alarm-server/logic/collector"
	"alarm-server/utils/modcall"
)

// BatchHandleMessage 批量处理告警/恢复消息
// 使用协程池处理kafka消息
func BatchHandleMessage(ctx context.Context, msgs []*sarama.ConsumerMessage) error {
	msgsList := make([][]*cpb.ValidateTaskItem, 0)
	for _, msg := range msgs {
		validMsg := &cpb.ValidateTaskList{}
		err := proto.Unmarshal(msg.Value, validMsg)
		if err != nil {
			log.Error("HandleMessage Unmarshal error: ", err)
			continue
		}
		msgsList = append(msgsList, validMsg.ValidTaskList)
	}
	AddToCollector(msgsList)
	return nil
}

// AddToCollector 添加到collector
func AddToCollector(msgsList [][]*cpb.ValidateTaskItem) {
	modcall.RecordRuleValidMsgCnt(len(msgsList))
	var realtimeMap = make(map[string]*model.ValidStoreData)
	var delaytimeMap = make(map[string]*model.ValidStoreData)
	for _, msgList := range msgsList {
		for _, record := range msgList {
			validStoreItem := &model.ValidStoreData{
				MozuId:      int32(record.MozuId),
				Rid:         record.Rid,
				Gid:         record.Gid,
				AlarmLevel:  record.AlarmLevel,
				EvalTime:    record.RunTime,
				PvTime:      record.RunTime,
				Success:     record.Successed,
				Fired:       record.Fired,
				ErrorCode:   int32(record.ErrorCode),
				ErrorName:   record.ErrorName,
				ErrorDetail: record.ErrorDetail,
			}
			if record.RidType == 0 {
				if validItem, ok := realtimeMap[validStoreItem.GetKey()]; ok {
					if validItem.EvalTime >= record.RunTime {
						continue
					}
				}
				realtimeMap[validStoreItem.GetKey()] = validStoreItem
			} else {
				if validItem, ok := delaytimeMap[validStoreItem.GetKey()]; ok {
					if validItem.EvalTime >= record.RunTime {
						continue
					}
				}
				delaytimeMap[validStoreItem.GetKey()] = validStoreItem
			}
		}
	}
	// 存储策略生效结果
	if len(realtimeMap) > 0 {
		collector.GetValidateColleror().BatchAddRuleRecord(realtimeMap, 0)
	}
	if len(delaytimeMap) > 0 {
		collector.GetValidateColleror().BatchAddRuleRecord(delaytimeMap, 1)
	}
}
