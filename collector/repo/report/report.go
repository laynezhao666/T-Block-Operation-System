// Package report 提供和指标上报相关的函数
package report

import (
	"collector/entity/config"

	"etrpc-go/log"

	"trpc.group/trpc-go/trpc-go/metrics"
)

const (
	metricGroupName string = "custom_report"
)

const (
	maxLatencyMetric         string = "max_latency"
	avgLatencyMetric         string = "avg_latency"
	handleCntMetric          string = "handle_cnt"
	handleFailCntMetric      string = "handle_fail_cnt"
	sendDataCntMetric        string = "send_data_cnt"
	sendDataFailCntMetric    string = "send_data_fail_cnt"
	fetchConfigCntMetric     string = "fetch_config_cnt"
	fetchConfigFailCntMetric string = "fetch_config_fail_cnt"
)

const (
	HandleTypeDimension  string = "handle_type"
	SendTypeDimension    string = "send_type"
	TargetDimension      string = "target"
	PointTypeDimension   string = "point_type"
	UpstreamIpDimension  string = "upstream_ip"
	FetchTypeDimension   string = "fetch_type"
	FetcherNameDimension string = "fetcher_name"
)

var (
	commonDimensions map[string]string
)

// Init 初始化指标上报
func Init() {
	commonDimensions = config.GetCommonReportConf().DimensionMap
	log.Infof("common dimension %v ", commonDimensions)
}

func report(dims []*metrics.Dimension, metric []*metrics.Metrics) error {
	// 上报前加上通用的维度
	for k, v := range commonDimensions {
		dims = append(dims, &metrics.Dimension{Name: k, Value: v})
	}
	return metrics.ReportMultiDimensionMetricsX(metricGroupName, dims, metric)
}

// HandleLatency 请求处理时延上报
func HandleLatency(dimensions []*metrics.Dimension, latency float64) {
	metric := []*metrics.Metrics{
		metrics.NewMetrics(maxLatencyMetric, latency, metrics.PolicyMAX),
		metrics.NewMetrics(avgLatencyMetric, latency, metrics.PolicyAVG),
	}
	err := report(dimensions, metric)
	if err != nil {
		log.Errorf("Report HandleLatency err %v, dimensions: %v", err, dimensions)
	}
}

// HandleCnt 请求处理数量上报
func HandleCnt(dimensions []*metrics.Dimension, cnt float64) {
	metric := []*metrics.Metrics{
		metrics.NewMetrics(handleCntMetric, cnt, metrics.PolicySUM),
	}
	err := report(dimensions, metric)
	if err != nil {
		log.Errorf("Report HandleCnt err %v, dimensions: %v", err, dimensions)
	}
}

// HandleFailCnt 请求处理失败数量上报
func HandleFailCnt(dimensions []*metrics.Dimension, cnt float64) {
	metric := []*metrics.Metrics{
		metrics.NewMetrics(handleFailCntMetric, cnt, metrics.PolicySUM),
	}
	err := report(dimensions, metric)
	if err != nil {
		log.Errorf("Report HandleFailCnt err %v, dimensions: %v", err, dimensions)
	}
}

func SendDataCnt(dimensions []*metrics.Dimension, cnt float64) {
	metric := []*metrics.Metrics{
		metrics.NewMetrics(sendDataCntMetric, cnt, metrics.PolicySUM),
	}
	err := report(dimensions, metric)
	if err != nil {
		log.Errorf("Report SendDataCnt err %v, dimensions: %v", err, dimensions)
	}
}

func SendDataFailCnt(dimensions []*metrics.Dimension, cnt float64) {
	metric := []*metrics.Metrics{
		metrics.NewMetrics(sendDataFailCntMetric, cnt, metrics.PolicySUM),
	}
	err := report(dimensions, metric)
	if err != nil {
		log.Errorf("Report SendDataFailCnt err %v, dimensions: %v", err, dimensions)
	}
}

func FetchConfigCnt(dimensions []*metrics.Dimension, cnt float64) {
	metric := []*metrics.Metrics{
		metrics.NewMetrics(fetchConfigCntMetric, cnt, metrics.PolicySUM),
	}
	err := report(dimensions, metric)
	if err != nil {
		log.Errorf("Report FetchConfigCnt err %v, dimensions: %v", err, dimensions)
	}
}

func FetchConfigFailCnt(dimensions []*metrics.Dimension, cnt float64) {
	metric := []*metrics.Metrics{
		metrics.NewMetrics(fetchConfigFailCntMetric, cnt, metrics.PolicySUM),
	}
	err := report(dimensions, metric)
	if err != nil {
		log.Errorf("Report FetchConfigFailCnt err %v, dimensions: %v", err, dimensions)
	}
}
