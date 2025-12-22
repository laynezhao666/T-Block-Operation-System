// Package modcall modcall
package modcall

import (
	"fmt"
	"time"

	"etrpc-go/log"

	"trpc.group/trpc-go/trpc-go"
	"trpc.group/trpc-go/trpc-go/metrics"
)

const (
	timerGroup string = "custom_report"

	//告警生效率相关
	DistinctTotalStrategyCnt string = "distinct_total_strategy_cnt"
	DistinctExecStrategyCnt  string = "distinct_exec_strategy_cnt"
	DistinctValidStrategyCnt string = "distinct_valid_strategy_cnt"
	RuleValidMsgCnt          string = "rule_valid_msg_cnt"
	WriteRedisDelay          string = "write_redis_delay"
	ReadRedisDelay           string = "read_redis_delay"
)

// RecordRuleValidMsgCnt 上报策略kafka消费数量
func RecordRuleValidMsgCnt(cnt int) {
	dimensions := []*metrics.Dimension{
		{Name: "strategy", Value: "TotalRT"},
		{Name: "set_name", Value: trpc.GlobalConfig().Global.FullSetName},
	}
	totalMetric := []*metrics.Metrics{
		metrics.NewMetrics(RuleValidMsgCnt, float64(cnt), metrics.PolicySUM),
	}
	err := metrics.ReportMultiDimensionMetricsX(timerGroup, dimensions, totalMetric)
	if err != nil {
		log.Errorf("Total RecordRuleValidMsgCnt err %v", err)
	}
}

// RecordDistinctAnalyzeCnt 上报分析策略数量
func RecordDistinctAnalyzeCnt(service string, mozuId, validCnt, execCnt, totalCnt int) {
	dimensions := []*metrics.Dimension{
		{Name: "strategy", Value: service},
		{Name: "mozu_id", Value: fmt.Sprintf("%v", mozuId)},
		{Name: "set_name", Value: trpc.GlobalConfig().Global.FullSetName},
	}
	totalMetric := []*metrics.Metrics{
		metrics.NewMetrics(DistinctTotalStrategyCnt, float64(totalCnt), metrics.PolicyMAX),
	}
	err := metrics.ReportMultiDimensionMetricsX(timerGroup, dimensions, totalMetric)
	if err != nil {
		log.Errorf("Total RecordDistinctAnalyzeCnt err %v, service %v", err, service)
	}
	execMetric := []*metrics.Metrics{
		metrics.NewMetrics(DistinctExecStrategyCnt, float64(execCnt), metrics.PolicyMAX),
	}
	err = metrics.ReportMultiDimensionMetricsX(timerGroup, dimensions, execMetric)
	if err != nil {
		log.Errorf("Exec RecordDistinctAnalyzeCnt err %v, service %v", err, service)
	}
	validMetric := []*metrics.Metrics{
		metrics.NewMetrics(DistinctValidStrategyCnt, float64(validCnt), metrics.PolicyMAX),
	}
	err = metrics.ReportMultiDimensionMetricsX(timerGroup, dimensions, validMetric)
	if err != nil {
		log.Errorf("Valid RecordDistinctAnalyzeCnt err %v, service %v", err, service)
	}
}

// RecordWriteRedisDelay 上报写redis延迟
func RecordWriteRedisDelay(delay time.Duration) {
	dimensions := []*metrics.Dimension{
		{Name: "strategy", Value: "TotalRT"},
		{Name: "set_name", Value: trpc.GlobalConfig().Global.FullSetName},
	}
	totalMetric := []*metrics.Metrics{
		metrics.NewMetrics(WriteRedisDelay, float64(delay.Milliseconds()), metrics.PolicyMAX),
	}
	err := metrics.ReportMultiDimensionMetricsX(timerGroup, dimensions, totalMetric)
	if err != nil {
		log.Errorf("Total RecordWriteRedisDelay err %v", err)
	}
}

// RecordReadRedisDelay 上报读redis延迟
func RecordReadRedisDelay(mozuId int64, delay time.Duration) {
	dimensions := []*metrics.Dimension{
		{Name: "strategy", Value: "TotalRT"},
		{Name: "mozu_id", Value: fmt.Sprintf("%v", mozuId)},
		{Name: "set_name", Value: trpc.GlobalConfig().Global.FullSetName},
	}
	totalMetric := []*metrics.Metrics{
		metrics.NewMetrics(ReadRedisDelay, float64(delay.Milliseconds()), metrics.PolicyMAX),
	}
	err := metrics.ReportMultiDimensionMetricsX(timerGroup, dimensions, totalMetric)
	if err != nil {
		log.Errorf("Total RecordReadRedisDelay err %v", err)
	}
}
