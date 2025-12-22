package rpc

import (
	"encoding/json"
	"sync"

	"etrpc-go/log"

	"github.com/IBM/sarama"
	"github.com/avast/retry-go"
	"trpc.group/trpc-go/trpc-database/kafka"
	"trpc.group/trpc-go/trpc-go"

	cmodel "common/entity/model"
)

const (
	cgiKafkaName = "trpc.kafka.tbos.cgi_sync"
	focKafkaName = "trpc.kafka.tbos.foc_sync"
)

var (
	cOnce  sync.Once
	client *KafkaClient
)

// KafkaClient kafka客户端
type KafkaClient struct {
	cgiCli kafka.Client
}

// GetCkafka 获取kafka客户端
func GetCkafka() *KafkaClient {
	cOnce.Do(func() {
		client = &KafkaClient{
			cgiCli: kafka.NewClientProxy(cgiKafkaName),
		}
	})
	return client
}

// SendCgiAlarm 发送cgi同步告警
func (c *KafkaClient) SendCgiAlarm(alarms []cmodel.AlarmActive) error {
	data, err := json.Marshal(alarms)
	if err != nil {
		log.Errorf("SendCgiAlarm json.Marshal err:%s", err.Error())
		return err
	}
	key := "CgiSynAlarm"
	retry.Do(func() error {
		_, _, err = c.cgiCli.SendSaramaMessage(trpc.BackgroundContext(), sarama.ProducerMessage{
			Key:   sarama.StringEncoder(key),
			Value: sarama.ByteEncoder(data),
		})
		return err
	}, retry.Attempts(3), retry.RetryIf(func(err error) bool {
		return err != nil
	}))
	if err != nil {
		log.Errorf("SendCgiAlarm SendSaramaMessage err:%s", err.Error())
		return err
	}
	return nil
}
