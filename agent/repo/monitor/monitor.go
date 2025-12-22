// Package monitor 智研监控
package monitor

import (
	"trpc.group/trpc-go/trpc-go/metrics"

	"agent/entity/config"

	"trpc.group/trpc-go/trpc-go/log"
)

const (
	reportName = "custom_report"
)

// Init 初始化
func Init() error {
	if !config.GetRB().IsFeatureEnable("zhiyan") {
		return nil
	}

	// todo 初始化公共维度信息，包含园区、模组
	return nil
}

// ReportMetricsWithDimensions 监控上报
func ReportMetricsWithDimensions(name string, value float64, policy metrics.Policy, dimensions []*metrics.Dimension) {
	m := []*metrics.Metrics{metrics.NewMetrics(name, value, policy)}
	err := metrics.ReportMultiDimensionMetricsX(reportName, dimensions, m)
	if err != nil {
		log.Errorf("Report Metrcs err:%s", err)
	}
}

// ReportMultiMetricsWithDimensions 监控上报
func ReportMultiMetricsWithDimensions(m []*metrics.Metrics, dimensions []*metrics.Dimension) {
	err := metrics.ReportMultiDimensionMetricsX(reportName, dimensions, m)
	if err != nil {
		log.Errorf("Report Multi Metrcs err:%s", err)
	}
}
