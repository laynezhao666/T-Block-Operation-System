package repo

import (
	"encoding/json"
	"sync"
	"time"

	"etrpc-go/log"
	pb "trpcprotocol/alarm-compute"

	"github.com/IBM/sarama"
	"github.com/avast/retry-go"
	"google.golang.org/protobuf/proto"
	"trpc.group/trpc-go/trpc-database/kafka"
	"trpc.group/trpc-go/trpc-go"

	"alarm-compute/utils/modcall"
)

const (
	ruleValidKafkaName = "trpc.kafka.tbos.rule"
	alertKafkaName     = "trpc.kafka.tbos.alert"
	dataKafkaName      = "trpc.kafka.tbos.data"
	adminKafkaName     = "trpc.kafka.tbos.admin"
)

var (
	cOnce  sync.Once
	client *KafkaClient
)

// KafkaClient kafka客户端
type KafkaClient struct {
	ruleCli  kafka.Client
	alertCli kafka.Client
	dataCli  kafka.Client
	adminCli kafka.Client
}

// GetCkafka 获取kafka客户端
func GetCkafka() *KafkaClient {
	cOnce.Do(func() {
		client = &KafkaClient{
			ruleCli:  kafka.NewClientProxy(ruleValidKafkaName),
			alertCli: kafka.NewClientProxy(alertKafkaName),
			dataCli:  kafka.NewClientProxy(dataKafkaName),
			adminCli: kafka.NewClientProxy(adminKafkaName),
		}
	})
	return client
}

// batchSendRuleValidMsg 批量通过kafka发送任务生效消息
func (c *KafkaClient) batchSendRuleValidMsg(recordList []*pb.ValidateTaskItem, batchSize int) {
	cnt := len(recordList)
	curList := []*pb.ValidateTaskItem{}
	i := 0
	for _, record := range recordList {
		i++
		curList = append(curList, record)
		if i%batchSize == 0 {
			c.SendRuleValidMsg(curList)
			curList = []*pb.ValidateTaskItem{}
		} else if i == cnt {
			if len(curList) > 0 {
				c.SendRuleValidMsg(curList)
			}
		}
	}
	modcall.RecordRuleValidKafkaCnt(cnt/batchSize + 1)
}

// SendRuleValidMsg 通过kafka发送任务生效消息
func (c *KafkaClient) SendRuleValidMsg(recordList []*pb.ValidateTaskItem) error {
	dataPb := &pb.ValidateTaskList{
		ValidTaskList: recordList,
	}
	data, err := proto.Marshal(dataPb)
	if err != nil {
		return err
	}
	_, _, err = c.ruleCli.SendSaramaMessage(trpc.BackgroundContext(), sarama.ProducerMessage{
		Key:   sarama.ByteEncoder([]byte("AlarmValidate")),
		Value: sarama.ByteEncoder(data),
	})
	if err != nil {
		log.Errorf("SendRuleValidMsg err %v", err)
	}
	return err
}

// SendAlertMsg 通过kafka发送告警消息
func (c *KafkaClient) SendAlertMsg(key []byte, data []byte) error {
	var err error
	retry.Do(func() error {
		_, _, err = c.alertCli.SendSaramaMessage(trpc.BackgroundContext(), sarama.ProducerMessage{
			Key:   sarama.ByteEncoder(key),
			Value: sarama.ByteEncoder(data),
		})
		return err
	}, retry.Attempts(3), retry.RetryIf(func(err error) bool {
		return err != nil
	}))
	return err
}

// SendPointMsg 通过kafka发送数据消息
func (c *KafkaClient) SendPointMsg(key []byte, data []byte) error {
	var err error
	retry.Do(func() error {
		_, _, err = c.dataCli.SendSaramaMessage(trpc.BackgroundContext(), sarama.ProducerMessage{
			Key:   sarama.ByteEncoder(key),
			Value: sarama.ByteEncoder(data),
		})
		return err
	}, retry.Attempts(3), retry.RetryIf(func(err error) bool {
		return err != nil
	}))
	return err
}

// Trace 运营平台链路数据
type Trace struct {
	TraceTimeType int32 `json:"et_type"` // 链路时间类型
	TraceTime     int64 `json:"et"`      // 链路时间戳，单位毫秒
}

// MQAdminPointMsgKey 运营平台测点消息队列的key
type MQAdminPointMsgKey struct {
	MozuId    string `json:"mID"`   // 模组ID
	DeviceID  string `json:"dID"`   // 设备ID
	WorkerID  string `json:"wID"`   // worker id, 生成uuid
	Seq       uint64 `json:"seq"`   // 序列号
	Timestamp int64  `json:"t"`     // 采集时间戳，单位秒
	Interval  int32  `json:"d"`     // 测点类型，60:周期性,1:变化测点
	PubMs     int64  `json:"pubMs"` // 投递kafka的毫秒时间戳，单位毫秒
	// 以下是运营平台扩展字段
	Traces []Trace `json:"traces"`
}

// SendAdminMsg 通过kafka向运营平台发送虚拟测点计算消息
// @param pubTime 虚拟测点数据投递kafka的时间
// @param data 虚拟测点数据
func (c *KafkaClient) SendAdminMsg(data []byte, calTime int64, pubTime time.Time) error {
	var err error
	msgKey := &MQAdminPointMsgKey{
		DeviceID:  "alarmVirtualPoints",
		Timestamp: calTime,
		Interval:  1,
		PubMs:     pubTime.UnixMilli(),
		Traces: []Trace{
			{
				TraceTimeType: 2,
				TraceTime:     pubTime.UnixMilli(),
			},
			{
				TraceTimeType: 3,
				TraceTime:     pubTime.UnixMilli(),
			},
		},
	}
	key, err := json.Marshal(msgKey)
	if err != nil {
		log.Errorf("marshal admin key msg failed, err: %v", err)
		return err
	}
	retry.Do(func() error {
		_, _, err = c.adminCli.SendSaramaMessage(trpc.BackgroundContext(), sarama.ProducerMessage{
			Key:   sarama.ByteEncoder(key),
			Value: sarama.ByteEncoder(data),
		})
		return err
	}, retry.Attempts(3), retry.RetryIf(func(err error) bool {
		return err != nil
	}))
	return err
}
