package utils

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"trpc.group/trpc-go/trpc-go/log"

	"github.com/segmentio/kafka-go/sasl/plain"

	"agent/entity/config"
	"agent/entity/definition"

	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/compress"
)

var (
	errorKafkaWriterNil = errors.New("kafka writer is nil")
)

// KafkaMessage kafka 消息类型
type KafkaMessage struct {
	Topic string
	Key   []byte
	Value []byte
}

// GetMessageKey 获取 kafka 消息 key
func GetMessageKey(deviceGid interface{}, sendTime time.Time, interval int) string {
	return fmt.Sprintf("%v###%v###%v", deviceGid, sendTime.Unix(), interval)
}

// NewKafkaWriter 创建 kafka writer
func NewKafkaWriter(brokers []string, balancer kafka.Balancer, topic string,
	mechanism string, username string, password string) *definition.KafkaWriterType {
	var kLogger kafka.Logger = nil
	if config.IsKafkaLogEnable() {
		kLogger = kafka.LoggerFunc(log.GetDefaultLogger().Infof)
	}

	var transport kafka.RoundTripper = nil
	switch strings.ToLower(mechanism) {
	case "plain":
		transport = &kafka.Transport{
			SASL: plain.Mechanism{
				Username: username,
				Password: password,
			},
		}
		break
	default:
		break
	}

	var compression = compress.None
	switch strings.ToLower(config.GetRB().Distributor.Kafka.Compression) {
	case "gzip":
		compression = kafka.Gzip
	case "snappy":
		compression = kafka.Snappy
	case "lz4":
		compression = kafka.Lz4
	case "zstd":
		compression = kafka.Zstd
	}

	return &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Balancer:     balancer,
		BatchBytes:   definition.KafkaMaxBatchBytes,
		MaxAttempts:  config.GetRB().Distributor.Kafka.MaxAttempt,
		WriteTimeout: time.Duration(config.GetRB().Distributor.Kafka.WriteTimeoutMs) * time.Millisecond,
		RequiredAcks: kafka.RequireOne,
		Compression:  compression,
		Logger:       kLogger,
		ErrorLogger:  kafka.LoggerFunc(log.GetDefaultLogger().Errorf),
		Transport:    transport,
	}
}

// WriteData 写数据
func WriteData(context context.Context, writer *definition.KafkaWriterType, messages ...KafkaMessage) error {
	if writer == nil {
		return errorKafkaWriterNil
	}
	if len(messages) == 0 {
		return nil
	}

	kafkaMessages := make([]kafka.Message, 0, len(messages))
	for _, msg := range messages {
		kafkaMessages = append(kafkaMessages, kafka.Message{
			Topic: msg.Topic,
			Key:   msg.Key,
			Value: msg.Value,
		})
	}
	return writer.WriteMessages(context, kafkaMessages...)
}
