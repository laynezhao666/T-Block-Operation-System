// Package ckafka 策略生效信息发送运营平台
package ckafka

import (
	"fmt"
	"sync"
	"time"

	"etrpc-go/log"

	"github.com/IBM/sarama"
	"github.com/avast/retry-go"
	"google.golang.org/protobuf/proto"
	"trpc.group/trpc-go/trpc-database/kafka"
	"trpc.group/trpc-go/trpc-go"

	"alarm-server/conf"
	"alarm-server/entity/model"

	pb "trpcprotocol/alarm-server"
)

const (
	// errDetailMsgTemplate 错误信息模板 error_name: error_msg
	errDetailMsgTemplate = "%s:%s"
	ruleAdminKafkaName   = "trpc.kafka.tbos.alarm_admin"
)

var (
	cOnce  sync.Once
	client *KafkaClient
)

// KafkaClient kafka客户端
type KafkaClient struct {
	ruleCli kafka.Client
}

// GetCkafka 获取kafka客户端
func GetCkafka() *KafkaClient {
	cOnce.Do(func() {
		client = &KafkaClient{
			ruleCli: kafka.NewClientProxy(ruleAdminKafkaName),
		}
	})
	return client
}

// sendAdminMsg 通过kafka发送任务生效消息
func (c *KafkaClient) sendAdminMsg(mozuId int32, list []*pb.MQStrategyMsgValue_Item) error {
	adminPbVal := &pb.MQStrategyMsgValue{
		List: list,
	}
	adminPbKey := &pb.MQStrategyMsgKey{
		MozuId: mozuId,
		List: []*pb.MQStrategyMsgKey_Trace{
			{
				EtType: 6, // 标识告警的上报时间
				Er:     time.Now().UnixMilli(),
			},
		},
	}
	key, _ := proto.Marshal(adminPbKey)
	data, _ := proto.Marshal(adminPbVal)
	var err error
	retry.Do(func() error {
		_, _, err = c.ruleCli.SendSaramaMessage(trpc.BackgroundContext(), sarama.ProducerMessage{
			Key:   sarama.ByteEncoder(key),
			Value: sarama.ByteEncoder(data),
		})
		return err
	}, retry.Attempts(3), retry.RetryIf(func(err error) bool {
		return err != nil
	}))
	if err != nil {
		log.Errorf("sendAdminMsg err %v", err)
	}
	return err
}

// BatchSendAdminRuleMsg 发送策略计算信息到kafka
// 分批发送
func (c *KafkaClient) BatchSendAdminRuleMsg(recordMap map[int64]map[string]*model.ValidStoreData) error {
	batchSize := int(conf.ServerConf.RuleValidConfig.BatchSize)
	if batchSize <= 0 {
		batchSize = 1000
	}
	curList := []*pb.MQStrategyMsgValue_Item{}
	i := 0
	var mozuId int32
	for _, gidMap := range recordMap {
		for _, record := range gidMap {
			i++
			mozuId = int32(record.MozuId)
			item := &pb.MQStrategyMsgValue_Item{
				Rid:       int32(record.Rid),
				DeviceGid: record.Gid,
				EvalTime:  record.EvalTime,
				PvTime:    record.PvTime,
				ErrorCode: int32(record.ErrorCode),
				Succeed:   record.Success,
				Fired:     record.Fired,
			}
			if !record.Success {
				item.ErrorDetail = fmt.Sprintf(errDetailMsgTemplate, record.ErrorName, record.ErrorDetail)
			}
			curList = append(curList, item)
			if i%batchSize == 0 {
				c.sendAdminMsg(mozuId, curList)
				curList = []*pb.MQStrategyMsgValue_Item{}
			}
		}
	}
	if len(curList) > 0 {
		c.sendAdminMsg(mozuId, curList)
	}
	return nil
}
