package modcall

import (
	"fmt"
	"strconv"

	"etrpc-go/log"

	"trpc.group/trpc-go/trpc-go"
	"trpc.group/trpc-go/trpc-go/metrics"
)

var (
	timerGroup string = "custom_report"
	// AlarmTimer
	AlarmTimer string = "alarm_compute"
	// Strategy_Time_Cost
	StrategyTimeCost string = "strategy_time_cost"
	// DataQuery
	DataQuery string = "data_query"
	// DataChange
	DataChange string = "data_change"
	// AnalyzeTaskCnt
	AnalyzeTaskCnt string = "analyze_task_cnt"
	// AnalyzeTaskInvalidCnt
	AnalyzeTaskInvalidCnt string = "analyze_task_invalid_cnt"

	RuleValidKafkaCnt string = "rule_valid_kafka_cnt"

	TotalAnalyzeCnt string = "total_analyze_cnt"
	DelayAnalyzeCnt string = "delay_analyze_cnt"
	ProduceAlertCnt string = "produce_alert_cnt"
)

func init() {
	alarmComputeHisOpt := metrics.HistogramOption{
		BucketBounds: []float64{1, 5, 10, 25, 50, 80, 100, 200, 500, 800, 1000},
	}
	metrics.RegisterHistogram(AlarmTimer, alarmComputeHisOpt)
	dataQueryHisOpt := metrics.HistogramOption{
		BucketBounds: []float64{1, 5, 10, 25, 50, 80, 100, 200, 500, 800, 1000},
	}
	metrics.RegisterHistogram(DataQuery, dataQueryHisOpt)
	dataChangeHisOpt := metrics.HistogramOption{
		BucketBounds: []float64{1, 5, 10, 25, 50, 80, 100, 200, 500, 800, 1000},
	}
	metrics.RegisterHistogram(DataChange, dataChangeHisOpt)
}

// RecordAlarmComputeTime 上报告警计算延时时间
func RecordAlarmComputeTime(service string, recordTime float64) {
	dimensions := []*metrics.Dimension{
		{Name: "strategy", Value: service},
		{Name: "set_name", Value: trpc.GlobalConfig().Global.FullSetName},
	}
	metric := []*metrics.Metrics{
		metrics.NewMetrics(AlarmTimer, recordTime, metrics.PolicyHistogram),
	}
	err := metrics.ReportMultiDimensionMetricsX(timerGroup, dimensions, metric)
	if err != nil {
		log.Errorf("RecordAlarmComputeTime err %v, service %v", err, service)
	}
}

// RecordDataQueryTime 上报数据查询延时时间
func RecordDataQueryTime(service string, qtype string, recordTime float64) {
	dimensions := []*metrics.Dimension{
		{Name: "strategy", Value: service},
		{Name: "query_type", Value: qtype},
		{Name: "set_name", Value: trpc.GlobalConfig().Global.FullSetName},
	}
	metric := []*metrics.Metrics{
		metrics.NewMetrics(DataQuery, recordTime, metrics.PolicyHistogram),
	}
	err := metrics.ReportMultiDimensionMetricsX(timerGroup, dimensions, metric)
	if err != nil {
		log.Errorf("RecordDataQueryTime err %v, service %v", err, service)
	}
}

// RecordDataChangeTime 上报变化测点延时时间
func RecordDataChangeTime(service string, recordTime float64) {
	dimensions := []*metrics.Dimension{
		{Name: "strategy", Value: service},
		{Name: "set_name", Value: trpc.GlobalConfig().Global.FullSetName},
	}
	metric := []*metrics.Metrics{
		metrics.NewMetrics(DataChange, recordTime, metrics.PolicyHistogram),
	}
	err := metrics.ReportMultiDimensionMetricsX(timerGroup, dimensions, metric)
	if err != nil {
		log.Errorf("RecordDataChangeTime err %v, service %v", err, service)
	}
}

// RecordAnalyzeTaskCnt 上报分析任务数
func RecordAnalyzeTaskCnt(service string, isTotal bool, cnt int) {
	var analyzeType string
	if isTotal {
		analyzeType = "全量分析"
	} else {
		analyzeType = "变化分析"
	}
	dimensions := []*metrics.Dimension{
		{Name: "strategy", Value: service},
		{Name: "analyze_type", Value: analyzeType},
		{Name: "set_name", Value: trpc.GlobalConfig().Global.FullSetName},
	}
	metric := []*metrics.Metrics{
		metrics.NewMetrics(AnalyzeTaskCnt, float64(cnt), metrics.PolicySUM),
	}
	err := metrics.ReportMultiDimensionMetricsX(timerGroup, dimensions, metric)
	if err != nil {
		log.Errorf("RecordAnalyzeTaskCnt err %v, service %v", err, service)
	}
}

// RecordInvalidAnalyzeTaskCnt 上报无效分析任务数
func RecordInvalidAnalyzeTaskCnt(service string, mozuId int, reason string, cnt int) {
	dimensions := []*metrics.Dimension{
		{Name: "strategy", Value: service},
		{Name: "mozu_id", Value: fmt.Sprintf("%v", mozuId)},
		{Name: "set_name", Value: trpc.GlobalConfig().Global.FullSetName},
	}
	metric := []*metrics.Metrics{
		metrics.NewMetrics(AnalyzeTaskInvalidCnt, float64(cnt), metrics.PolicySUM),
	}
	err := metrics.ReportMultiDimensionMetricsX(timerGroup, dimensions, metric)
	if err != nil {
		log.Errorf("RecordInvalidAnalyzeTaskCnt err %v, service %v", err, service)
	}
}

// RecordStrategyTimeCost 上报策略计算时延
func RecordStrategyTimeCost(service string, isTotal bool, recordTime float64) {
	var analyzeType string
	if isTotal {
		analyzeType = "全量分析"
	} else {
		analyzeType = "变化分析"
	}
	dimensions := []*metrics.Dimension{
		{Name: "strategy", Value: service},
		{Name: "analyze_type", Value: analyzeType},
		{Name: "set_name", Value: trpc.GlobalConfig().Global.FullSetName},
	}
	metric := []*metrics.Metrics{
		metrics.NewMetrics(StrategyTimeCost, recordTime, metrics.PolicyAVG),
	}
	err := metrics.ReportMultiDimensionMetricsX(timerGroup, dimensions, metric)
	if err != nil {
		log.Errorf("RecordInvalidAnalyzeTaskCnt err %v, service %v", err, service)
	}
}

// RecordRuleValidKafkaCnt 上报规则生效kafka消息数量
func RecordRuleValidKafkaCnt(cnt int) {
	dimensions := []*metrics.Dimension{
		{Name: "strategy", Value: "All"},
		{Name: "set_name", Value: trpc.GlobalConfig().Global.FullSetName},
	}
	metric := []*metrics.Metrics{
		metrics.NewMetrics(RuleValidKafkaCnt, float64(cnt), metrics.PolicySUM),
	}
	err := metrics.ReportMultiDimensionMetricsX(timerGroup, dimensions, metric)
	if err != nil {
		log.Errorf("RecordAnalyzeTaskCnt err %v", err)
	}
}

// RecordAnalyzeDelayCnt 上报分析延迟次数
func RecordAnalyzeDelayCnt(service string, isDelay bool) {
	dimensions := []*metrics.Dimension{
		{Name: "strategy", Value: service},
		{Name: "set_name", Value: trpc.GlobalConfig().Global.FullSetName},
	}
	metric := []*metrics.Metrics{
		metrics.NewMetrics(TotalAnalyzeCnt, 1, metrics.PolicySUM),
	}
	if isDelay {
		metric = append(metric,
			metrics.NewMetrics(DelayAnalyzeCnt, 1, metrics.PolicySUM))
	}
	err := metrics.ReportMultiDimensionMetricsX(timerGroup, dimensions, metric)
	if err != nil {
		log.Errorf("RecordAnalyzeDelayCnt err %v, service %v", err, service)
	}
}

// RecordProduceAlertCnt 上报生产的告警消息数量
func RecordProduceAlertCnt(mozuId, cnt int) {
	dimensions := []*metrics.Dimension{
		{Name: "mozu_id", Value: strconv.FormatInt(int64(mozuId), 10)},
		{Name: "set_name", Value: trpc.GlobalConfig().Global.FullSetName},
	}
	totalMetric := []*metrics.Metrics{
		metrics.NewMetrics(ProduceAlertCnt, float64(cnt), metrics.PolicySUM),
	}
	err := metrics.ReportMultiDimensionMetricsX(timerGroup, dimensions, totalMetric)
	if err != nil {
		log.Errorf("RecordProduceAlertCnt err %v", err)
	}
}
