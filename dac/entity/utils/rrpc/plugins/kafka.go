// Package plugins 提供RRPC消息代理的插件实现。
package plugins

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/segmentio/kafka-go"
)

// kafkaFieldNum Kafka消息Key中的字段数量
const (
	kafkaFieldNum = 4
)

// kafkaProxy 基于Kafka的RRPC消息代理
type kafkaProxy struct {
	writer *kafka.Writer
	reader *kafka.Reader
	ctx    context.Context
}

// NewKafkaProxy 创建Kafka消息代理实例，自动创建所需的Topic
func NewKafkaProxy(ctx context.Context,
	brokers []string, reqTopic, respTopic string,
) (*kafkaProxy, error) {
	if len(brokers) == 0 {
		return nil, errors.New("empty brokers")
	}
	if err := createTopics(
		ctx, brokers[0], reqTopic, respTopic,
	); err != nil {
		return nil, err
	}

	return &kafkaProxy{
		writer: &kafka.Writer{
			Addr:  kafka.TCP(brokers...),
			Topic: reqTopic,
		},
		reader: kafka.NewReader(
			kafka.ReaderConfig{
				Brokers:     brokers,
				GroupID:     "rrpc_resp",
				GroupTopics: nil,
				Topic:       respTopic,
			},
		),
		ctx: ctx,
	}, nil
}

// Write 向Kafka发送RRPC请求消息
func (k *kafkaProxy) Write(
	id, msgID string, payload []byte,
) error {
	key := []byte(fmt.Sprintf(
		"$SYS/rrpc_req/%v/%v", id, msgID))
	return k.writer.WriteMessages(
		context.Background(), kafka.Message{
			Key:   key,
			Value: payload,
		},
	)
}

// Read 从Kafka读取RRPC响应消息，返回消息ID和负载
func (k *kafkaProxy) Read() (string, []byte, error) {
	msg, err := k.reader.ReadMessage(k.ctx)
	if err != nil {
		return "", nil, err
	}

	key := string(msg.Key)
	fields := strings.Split(key, "/")
	l := len(fields)
	if l != kafkaFieldNum ||
		strings.HasPrefix(key, "$SYS/rrpc_req/") {
		return "", nil, fmt.Errorf(
			"error kafka message key: %v", key)
	}
	msgID := fields[l-1]
	return msgID, msg.Value, nil
}

// Close 关闭Kafka消费者连接
func (k *kafkaProxy) Close() error {
	return k.reader.Close()
}

// createTopics 在Kafka集群中创建指定的Topic
func createTopics(ctx context.Context,
	broker string, topics ...string,
) error {
	c, err := kafka.Dial("tcp", broker)
	if err != nil {
		return err
	}
	defer func() {
		_ = c.Close()
	}()

	controller, err := c.Controller()
	if err != nil {
		return err
	}

	addr := net.JoinHostPort(
		controller.Host, strconv.Itoa(controller.Port))
	conn, err := kafka.DialContext(ctx, "tcp", addr)
	if err != nil {
		return err
	}
	defer func() {
		_ = conn.Close()
	}()

	for _, topic := range topics {
		if err = conn.CreateTopics(
			kafka.TopicConfig{
				Topic:             topic,
				NumPartitions:     2,
				ReplicationFactor: 2,
			},
		); err != nil {
			return err
		}
	}

	return nil
}
