// Package report 指标上报相关定义
package report

import (
	"etrpc-go/metric"

	"github.com/samber/lo"
	"trpc.group/trpc-go/trpc-go/metrics"
)

const (
	MozuIdDimKey        = "mozu_id"
	KafkaSourceDimKey   = "source"
	PointTypeDimKey     = "point_type"
	PointIntervalDimKey = "interval"
)

var (
	KafkaConsumeDelay = metric.NewMetric("kafka_consume_delay", // Kafka消费时延
		metric.WithPolicy(metrics.PolicyAVG, metrics.PolicyMAX))
	KafkaConsumeMsgCnt   = metric.NewMetric("kafka_consume_msg_cnt")   // Kafka消费-总消息量
	KafkaConsumePointCnt = metric.NewMetric("kafka_consume_point_cnt") // Kafka消费-总测点量

	InfluxWriteSuccessCnt = metric.NewMetric("influx_write_success_cnt") // InfluxDB测点存储-成功数
	InfluxWriteFailCnt    = metric.NewMetric("influx_write_fail_cnt")    // InfluxDB测点存储-失败数
	InfluxWriteCost       = metric.NewMetric("influx_write_cost",        // InfluxDB测点存储耗时
		metric.WithPolicy(metrics.PolicyAVG, metrics.PolicyMAX))

	KafkaWriteSuccessCnt = metric.NewMetric("kafka_write_success_cnt") // Kafka测点存储-成功数
	KafkaWriteFailCnt    = metric.NewMetric("Kafka_write_fail_cnt")    // Kafka测点存储-失败数
	KafkaWriteCost       = metric.NewMetric("kafka_write_cost",        // Kafka测点存储耗时
		metric.WithPolicy(metrics.PolicyAVG, metrics.PolicyMAX))
)

func CountByMozu[T any](metric metric.IMetric, data []T, mozuIdFunc func(item T) int32) {
	mozuData := lo.GroupBy(data, mozuIdFunc)
	for mozuId, dt := range mozuData {
		metric.ReportWithDim(float64(len(dt)), map[string]string{MozuIdDimKey: string(mozuId)})
	}
}

func ValByMozu[T any](metric metric.IMetric, data []T, mozuIdFunc func(item T) int32, val float64) {
	mozuData := lo.GroupBy(data, mozuIdFunc)
	for mozuId := range mozuData {
		metric.ReportWithDim(val, map[string]string{MozuIdDimKey: string(mozuId)})
	}
}
