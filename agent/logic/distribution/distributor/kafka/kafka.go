// Package kafka kafka转发
package kafka

import (
	"context"
	"encoding/json"

	"trpc.group/trpc-go/trpc-go/log"

	"agent/entity/config"
	"agent/entity/consts"
	"agent/entity/definition"
	kafkaEntity "agent/entity/kafka"
	model2 "agent/entity/model/data"
	"agent/logic/cm"
	utils2 "agent/logic/distribution/distributor/utils"
	monitor2 "agent/repo/monitor"
	"agent/utils"
)

var (
	kafkaDt kafkaDistributor
)

// Init 初始化
func Init() error {
	kafkaEntity.Init()

	var err error
	if err = InitDistributors(); err != nil {
		return err
	}

	return nil
}

// UnInit 反初始化
func UnInit() {
	UnInitDistributors()
}

// InitDistributors 初始化分发器
func InitDistributors() error {
	kafkaDt = kafkaDistributor{
		writer:         kafkaEntity.GetWriter(),
		forwardsWriter: kafkaEntity.GetForwardsWriter(),
	}

	return nil
}

// UnInitDistributors 反初始化分发器
func UnInitDistributors() {
	_ = kafkaDt.writer.Close()
	for i := range kafkaDt.forwardsWriter {
		_ = kafkaDt.forwardsWriter[i].Close()
	}
}

type kafkaDistributor struct {
	writer         *definition.KafkaWriterType
	forwardsWriter []*definition.KafkaWriterType
}

// KafkaDistributor 分发器
func KafkaDistributor() *kafkaDistributor {
	return &kafkaDt
}

// NewKafkaDistributor 新建分发器
func NewKafkaDistributor(writer *definition.KafkaWriterType) *kafkaDistributor {
	return &kafkaDistributor{writer: writer}
}

func (k *kafkaDistributor) getTopic(deviceGiD definition.DeviceGidType) string {
	topic := config.GetRB().Distributor.Kafka.Topic.Points
	if config.GetRB().IsFeatureEnable(consts.EnableMozuTopic) {
		mozuID := cm.Worker().GetDeviceMozuID(deviceGiD)
		if len(mozuID) > 0 {
			topic += "_" + mozuID
		} else {
			log.Warnf("kafkaDistributor.Distribute getTopic error, DeviceGiD: %+v", deviceGiD)
		}
	}
	return topic
}

// Distribute 分发
func (k *kafkaDistributor) Distribute(data *model2.DataUnit, args ...interface{}) {
	if k == nil || data == nil || len(data.Points) == 0 {
		return
	}

	topic := k.getTopic(data.DeviceGid)

	sendTime, interval := utils2.GetSendTimeAndInterval(args)
	isDefault := utils2.IsDefaultInterval(interval)

	kData, kafkaDataList, err := utils2.ToKafkaData(data, interval, false)
	if err != nil {
		log.Errorf("kafkaDistributor.Distribute %+v error: %+v", data, err)
		return
	}
	if len(kData.Points) == 0 && len(kData.VirtualPoints) == 0 {
		return
	}

	utils2.DebugRecord(kData)

	messageLen := 0
	messages := make([]utils.KafkaMessage, 0, len(kafkaDataList))
	for _, kafkaData := range kafkaDataList {
		b, err := json.Marshal(kafkaData)
		if err != nil {
			log.Errorf("kafkaDistributor.Distribute Marshal %+v error: %+v", kafkaData, err)
			return
		}
		key := utils2.GetMessageKey(data.DeviceGid, sendTime.Unix(), interval)
		messageLen += len(b)
		messages = append(messages, utils.KafkaMessage{
			Topic: topic,
			Key:   []byte(key),
			Value: b})
	}

	// 转发测点数据到其他 kafka
	go forwardMessages(k.forwardsWriter, messages)

	shouldLog := utils2.GetLogger(data.DeviceGid).Insert(interval)

	err = utils.WriteData(context.Background(), k.writer, messages...)
	if err != nil {
		s := 0
		for i := range messages {
			s += len(messages[i].Value)
		}
		log.Warnf("write date error: %v, message num: %v, total bytes: %v", err, len(messages), s)

		if shouldLog {
			monitor2.LogReportError(err)
			log.Errorf(
				"kafka Distribute data: %+v, args: %+v, error: %v, retry http...", data.DeviceGid, args, err,
			)
		}
		retry(data, messages, kData, kafkaDataList, interval, shouldLog, isDefault)
	}

	if shouldLog {
		kData.Log("Kafka", data.DeviceGid, interval)
		log.Infof(
			"kafka message number: %+v, total bytes: %+v, push interval: %+v, device: %v", len(messages),
			messageLen, interval,
			data.DeviceGid,
		)
	}
	kData.Report("kafka", interval)
}
