// Package service 各种接口的处理入口
package service

import (
	"context"
	"data-cache/entity/consts"
	"data-cache/logic/consumer"

	"github.com/IBM/sarama"
)

var majorKafkaConsumer = consumer.NewPointKafkaConsumer(consts.TbosMajorKafkaName)
var backupKafkaConsumer = consumer.NewPointKafkaConsumer(consts.TbosBackupKafkaName)

// MajorKafkaHandle 测点kafka处理函数(主用)
func MajorKafkaHandle(ctx context.Context, msgs []*sarama.ConsumerMessage) error {
	return majorKafkaConsumer.BatchHandle(ctx, msgs)
}

// BackupKafkaHandle 测点kafka处理函数(备用)
func BackupKafkaHandle(ctx context.Context, msgs []*sarama.ConsumerMessage) error {
	return backupKafkaConsumer.BatchHandle(ctx, msgs)
}
