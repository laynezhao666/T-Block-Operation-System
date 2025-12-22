// Package report 指标上报相关逻辑
package report

import (
	"etrpc-go/metric"

	"trpc.group/trpc-go/trpc-go/metrics"
)

const (
	MozuIdDimKey        = "mozu_id"
	KafkaSourceDimKey   = "source"
	PointTypeDimKey     = "point_type"
	PointIntervalDimKey = "interval"
)

var (
	KafkaConsumeDelay = metric.NewMetric("kafka_consume_delay", // Kafka消费-时延
		metric.WithPolicy(metrics.PolicyAVG, metrics.PolicyMAX))
	KafkaConsumeMsgCnt   = metric.NewMetric("kafka_consume_msg_cnt")   // Kafka消费-总消息量
	KafkaConsumePointCnt = metric.NewMetric("kafka_consume_point_cnt") // Kafka消费-总测点量

	LocalWriteCnt  = metric.NewMetric("local_write_cnt")  // 本地测点存储-数量
	LocalWriteCost = metric.NewMetric("local_write_cost", // 本地测点存储-耗时
		metric.WithPolicy(metrics.PolicyAVG, metrics.PolicyMAX))

	LocalReadCnt  = metric.NewMetric("local_read_cnt")  // 本地测点读取-数量
	LocalReadCost = metric.NewMetric("local_read_cost", // 本地测点读取-耗时
		metric.WithPolicy(metrics.PolicyAVG, metrics.PolicyMAX))

	LocalReadChangedCnt  = metric.NewMetric("local_read_changed_cnt")  // 本地变化测点读取-数量
	LocalReadChangedCost = metric.NewMetric("local_read_changed_cost", // 本地变化测点读取-耗时
		metric.WithPolicy(metrics.PolicyAVG, metrics.PolicyMAX))
)
