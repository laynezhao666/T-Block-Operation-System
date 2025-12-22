package std

import (
	"agent/repo/monitor"

	"agent/entity/consts"

	rtdbModel "agent/logic/collector/rtdb/model"

	"trpc.group/trpc-go/trpc-go/log"
	"trpc.group/trpc-go/trpc-go/metrics"
)

func getStdDimensions() (dimensions []*metrics.Dimension) {
	attrsList := []*metrics.Dimension{
		{
			Name:  consts.AttrDeviceName,
			Value: "std",
		},
	}

	return attrsList
}

func (cal *calculator) reportStdQuaMetrics(sucess float64, changedAll float64) {
	m := []*metrics.Metrics{
		metrics.NewMetrics("std_changed_success", sucess, metrics.PolicySUM),
		metrics.NewMetrics("std_changed_all", changedAll, metrics.PolicySUM),
	}
	monitor.ReportMultiMetricsWithDimensions(m, getStdDimensions())
}

func (cal *calculator) reportStdTimeMetrics(duration float64) {
	m := []*metrics.Metrics{metrics.NewMetrics("std_changed_compute_duration", duration, metrics.PolicyAVG)}
	monitor.ReportMultiMetricsWithDimensions(m, getStdDimensions())
}

func (cal *calculator) quaAnalysisReport(points []rtdbModel.DataPoint) {
	quaStat := map[consts.Quality]int{}
	errList := []string{}
	for _, v := range points {
		qua := v.Rtd.Val.Qua
		if qua != consts.QualityOk {
			quaStat[qua]++
			errList = append(errList, string(v.ID))
		}
	}
	if len(quaStat) > 0 {
		log.Debugf("std qua stat:%+v", quaStat)
		if len(errList) > 20 {
			errList = errList[:20]
		}
		log.Debugf("std qua err points sample:%v", errList)
	}
	// todo 上报智研指标
}
