// Package modcall ..
package modcall

import (
	"fmt"

	"etrpc-go/log"

	"trpc.group/trpc-go/trpc-go"
	"trpc.group/trpc-go/trpc-go/metrics"
)

const (
	timerGroup        = "custom_report"
	ConsumeAlertCnt   = "consume_alert_cnt"    // 消费告警kafka消息数量
	TotalDBReqCnt     = "total_db_req_cnt"     // db请求总数
	SuccessDBReqCnt   = "success_db_req_cnt"   // db请求成功数
	TotalDBWriteCnt   = "total_db_write_cnt"   // db写入数
	SuccessDBWriteCnt = "success_db_write_cnt" // db成功写入数
	TotalFocPushCnt   = "total_foc_push_cnt"   // foc kafka推送
	SuccessFocPushCnt = "success_foc_push_cnt" // foc kafka推送成功数
)

// RecordConsumeAlertCnt 上报消费告警数量
func RecordConsumeAlertCnt(mozuId, cnt int) {
	dimensions := []*metrics.Dimension{
		{Name: "mozu_id", Value: fmt.Sprintf("%d", mozuId)},
		{Name: "set_name", Value: trpc.GlobalConfig().Global.FullSetName},
	}
	totalMetric := []*metrics.Metrics{
		metrics.NewMetrics(ConsumeAlertCnt, float64(cnt), metrics.PolicySUM),
	}
	err := metrics.ReportMultiDimensionMetricsX(timerGroup, dimensions, totalMetric)
	if err != nil {
		log.Errorf("RecordConsumeAlertCnt err %v", err)
	}
}

// RecordDBReqCnt 上报db请求数量 (成功 全体),监测数据库状态
func RecordDBReqCnt(success bool) {
	dimensions := []*metrics.Dimension{
		{Name: "set_name", Value: trpc.GlobalConfig().Global.FullSetName},
	}
	totalMetric := []*metrics.Metrics{
		metrics.NewMetrics(TotalDBReqCnt, float64(1), metrics.PolicySUM),
	}
	err := metrics.ReportMultiDimensionMetricsX(timerGroup, dimensions, totalMetric)
	if err != nil {
		log.Errorf("Total RecordDBReqCnt err %v", err)
	}
	if !success {
		return
	}
	successMetric := []*metrics.Metrics{
		metrics.NewMetrics(SuccessDBReqCnt, float64(1), metrics.PolicySUM),
	}
	err = metrics.ReportMultiDimensionMetricsX(timerGroup, dimensions, successMetric)
	if err != nil {
		log.Errorf("Success RecordDBReqCnt err %v", err)
	}
}

// RecordDBWriteCnt 上报db写入数量
func RecordDBWriteCnt(mozuId, cnt int32, success bool) {
	dimensions := []*metrics.Dimension{
		{Name: "mozu_id", Value: fmt.Sprintf("%d", mozuId)},
		{Name: "set_name", Value: trpc.GlobalConfig().Global.FullSetName},
	}
	totalMetric := []*metrics.Metrics{
		metrics.NewMetrics(TotalDBWriteCnt, float64(cnt), metrics.PolicySUM),
	}
	err := metrics.ReportMultiDimensionMetricsX(timerGroup, dimensions, totalMetric)
	if err != nil {
		log.Errorf("Total RecordDBWriteCnt err %v", err)
	}
	if !success {
		return
	}
	successMetric := []*metrics.Metrics{
		metrics.NewMetrics(SuccessDBWriteCnt, float64(cnt), metrics.PolicySUM),
	}
	err = metrics.ReportMultiDimensionMetricsX(timerGroup, dimensions, successMetric)
	if err != nil {
		log.Errorf("Success RecordDBWriteCnt err %v", err)
	}
}

// RecordFocPushCnt 上报Foc推送数量
func RecordFocPushCnt(mozuId, cnt int32, success bool) {
	dimensions := []*metrics.Dimension{
		{Name: "mozu_id", Value: fmt.Sprintf("%d", mozuId)},
		{Name: "set_name", Value: trpc.GlobalConfig().Global.FullSetName},
	}
	totalMetric := []*metrics.Metrics{
		metrics.NewMetrics(TotalFocPushCnt, float64(cnt), metrics.PolicySUM),
	}
	err := metrics.ReportMultiDimensionMetricsX(timerGroup, dimensions, totalMetric)
	if err != nil {
		log.Errorf("Total RecordFocPushCnt err %v", err)
	}
	if !success {
		return
	}
	successMetric := []*metrics.Metrics{
		metrics.NewMetrics(SuccessFocPushCnt, float64(cnt), metrics.PolicySUM),
	}
	err = metrics.ReportMultiDimensionMetricsX(timerGroup, dimensions, successMetric)
	if err != nil {
		log.Errorf("Success RecordFocPushCnt err %v", err)
	}
}
