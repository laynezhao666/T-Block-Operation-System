// Package virtualpoints 提供门禁控制器虚拟测点的采集和上报功能。
package virtualpoints

import (
	"dac/entity/consts"

	"dac/entity/utils/tmetrics"
	"trpc.group/trpc-go/trpc-go/metrics"
)

// reportMetric 异步上报单个指标
func (v *VirtualPoints) reportMetric(
	name string, value float64, policy metrics.Policy,
) {
	go func() {
		tmetrics.ReportMetricWithDimensions(name, value, policy, v.deviceAttrList)
	}()
}

// reportMetrics 异步批量上报指标
func (v *VirtualPoints) reportMetrics(m []*metrics.Metrics) {
	go func() {
		tmetrics.ReportMetricsWithDimensions(m, v.deviceAttrList)
	}()
}

// reportRequestCount 上报请求计数指标（总请求数和成功请求数）
func (v *VirtualPoints) reportRequestCount(success bool) {
	m := make([]*metrics.Metrics, 0, 2)
	m = append(m, metrics.NewMetrics(consts.IntervalIDTotalRequestCount, 1, metrics.PolicySUM))
	if success {
		m = append(m, metrics.NewMetrics(consts.IntervalIDSuccessRequestCount, 1, metrics.PolicySUM))
	}

	v.reportMetrics(m)
	v.reportMetrics(m)
}

// ReportComm 上报通讯状态指标
func (v *VirtualPoints) ReportComm(commState bool) {
	value := 0.0
	if commState {
		value = 1.0
	}
	v.reportMetric(consts.StandardIDCommunicationState, value, metrics.PolicySUM)
}
