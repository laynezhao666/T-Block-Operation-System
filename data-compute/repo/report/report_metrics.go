// Package report 定义一些需要上报的指标
package report

import (
	"etrpc-go/metric"

	"trpc.group/trpc-go/trpc-go/metrics"
)

const (
	CalcTypeKey         = "计算类型" // 计算类型Key
	PointIntervalDimKey = "interval"
	PointMozuIdDimKey   = "mozu_id"
	PointSourceDimKey   = "source"
)

var (
	PointCalcExpectCnt  = metric.NewMetric("point_calc_expect_cnt")  // 测点计算-期望数量
	PointCalcSuccessCnt = metric.NewMetric("point_calc_success_cnt") // 测点计算-成功数量
	PointCalcCost       = metric.NewMetric("point_calc_cost",        // 测点计算-耗时
		metric.WithPolicy(metrics.PolicyMAX, metrics.PolicyAVG))

	PointPushSuccessCnt = metric.NewMetric("point_push_success_cnt") // 测点推送-成功数量
	PointPushFailCnt    = metric.NewMetric("point_push_fail_cnt")    // 测点推送-失败数量
	PointPushCost       = metric.NewMetric("point_push_cost",        // 测点推送-耗时
		metric.WithPolicy(metrics.PolicyMAX, metrics.PolicyAVG))
)
