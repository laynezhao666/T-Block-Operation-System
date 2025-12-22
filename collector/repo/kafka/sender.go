// Package kafka 保存kafka相关函数，用于管理生产者以及发送消息
package kafka

import (
	"collector/entity/config"
	"context"
	"fmt"
	"strings"
	"sync"

	"etrpc-go/log"

	"trpc.group/trpc-go/trpc-database/kafka"
)

// KafkaSender kafka发送器，发送到kafka
type KafkaSender struct {
	cli  kafka.Client
	name string
}

type senderManager struct {
	mutex   sync.RWMutex
	senders map[string]*KafkaSender
}

const (
	kafkaServicePrefix        string = "trpc.kafka.producer."
	prefixLen                 int    = len(kafkaServicePrefix)
	mainStdKafkaServiceName   string = "trpc.kafka.producer.mainStdService"
	backupStdKafkaServiceName string = "trpc.kafka.producer.backupStdService"
	traceKafkaServiceName     string = "trpc.kafka.producer.traceService"
)

var (
	manager      *senderManager
	mainSender   *KafkaSender
	backupSender *KafkaSender
	traceSender  *KafkaSender
)

// Init 初始化
func Init() {
	manager = &senderManager{
		senders: make(map[string]*KafkaSender),
	}
	clientConf := config.GetClientConf()
	manager.mutex.Lock()
	defer manager.mutex.Unlock()
	for _, s := range clientConf.Service {
		name := s.Name
		if !strings.HasPrefix(name, kafkaServicePrefix) {
			continue
		}
		kafkaName := strings.TrimPrefix(name, kafkaServicePrefix)
		sender := &KafkaSender{
			name: kafkaName,
			cli:  kafka.NewClientProxy(name),
		}
		manager.senders[kafkaName] = sender
	}
	log.Info("kafka producer register done")
}

func SenderManager() *senderManager {
	return manager
}

func (s *senderManager) GetSenderByName(name string) *KafkaSender {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	if sender, ok := s.senders[name]; ok {
		return sender
	}
	return nil
}

// Send 将消息转发到Kafka
func (s *KafkaSender) Send(key []byte, value []byte) error {
	if s == nil {
		return fmt.Errorf("kafka sender is nil")
	}
	return s.cli.Produce(context.Background(), key, value)
}
