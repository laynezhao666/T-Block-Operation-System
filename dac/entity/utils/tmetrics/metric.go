package tmetrics

import (
	"trpc.group/trpc-go/trpc-go/metrics"
)

// ReportMetricWithDimensions 上报单个指标（带维度）
func ReportMetricWithDimensions(name string, value float64, policy metrics.Policy, dimensions []*metrics.Dimension) {
	_ = metrics.ReportMultiDimensionMetricsX(name, dimensions, []*metrics.Metrics{
		metrics.NewMetrics(name, value, policy),
	})
}

// ReportMetricsWithDimensions 上报多个指标（带维度）
func ReportMetricsWithDimensions(m []*metrics.Metrics, dimensions []*metrics.Dimension) {
	if len(m) == 0 {
		return
	}
	_ = metrics.ReportMultiDimensionMetricsX(m[0].Name(), dimensions, m)
}
