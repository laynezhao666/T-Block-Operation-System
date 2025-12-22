package virtualpoints

import (
	"agent/entity/consts"
	"agent/entity/definition"
	"agent/logic/cm"
	model2 "agent/logic/collector/rtdb/model"
	"agent/repo/monitor"
	"agent/utils"

	"trpc.group/trpc-go/trpc-go/metrics"
)

func (v *VirtualPoints) reportMetrics(name string, value float64, policy metrics.Policy) {
	m := []*metrics.Metrics{metrics.NewMetrics(name, value, policy)}
	monitor.ReportMultiMetricsWithDimensions(m, v.deviceAttrList)
}

func (v *VirtualPoints) reportChannelMetrics(channelIndex int, name string, value float64, policy metrics.Policy) {
	m := []*metrics.Metrics{metrics.NewMetrics(name, value, policy)}
	monitor.ReportMultiMetricsWithDimensions(m, v.channelAttrList[channelIndex])
}

func (v *VirtualPoints) reportMetricsList(m []*metrics.Metrics) {
	monitor.ReportMultiMetricsWithDimensions(m, v.deviceAttrList)

}

func (v *VirtualPoints) reportChannelMetricsList(channelIndex int, m []*metrics.Metrics) {
	monitor.ReportMultiMetricsWithDimensions(m, v.channelAttrList[channelIndex])
}

func (v *VirtualPoints) reportRequestCount(success bool, totalMessageCount uint64, successMessageCount uint64) {
	successValue := 0.0
	if success {
		successValue = 1.0
	}
	m := []*metrics.Metrics{
		metrics.NewMetrics(definition.TotalRequestCountID, 1, metrics.PolicySUM),
		metrics.NewMetrics(definition.SuccessRequestCountID, successValue, metrics.PolicySUM),
		metrics.NewMetrics(definition.TotalRequestMessageCountID, float64(totalMessageCount), metrics.PolicySET),
		metrics.NewMetrics(definition.SuccessRequestMessageCountID, float64(successMessageCount), metrics.PolicySET),
	}
	v.reportMetricsList(m)
}

// ReportInterruption 上报中断
func (v *VirtualPoints) ReportInterruption(interrupted bool) {
	value := 0.0
	if interrupted {
		value = 1.0
	}
	v.reportMetrics(definition.InterruptionID, value, metrics.PolicySUM)
}

// ReportChannelInterruption 上报中断
func (v *VirtualPoints) ReportChannelInterruption(channelIndex int, interrupted bool) {
	value := 0.0
	if interrupted {
		value = 1.0
	}
	v.reportChannelMetrics(channelIndex, definition.InterruptionID, value, metrics.PolicySUM)
}

func (v *VirtualPoints) reportTimeoutRequest() {
	v.reportMetrics(definition.TimeoutRequestCountID, 1, metrics.PolicySUM)
}

func (v *VirtualPoints) reportPointThroughput() {
	v.reportMetrics(definition.PointThroughputID, float64(v.onePeriodPointThroughput), metrics.PolicySET)
}

func (v *VirtualPoints) reportPeriodResponseTime() {
	v.reportMetrics(definition.TotalResponseTimeID, float64(v.onePeriodCostTimeMs), metrics.PolicySET)
}

// ReportDataPointMetrics 上报数据点
func ReportDataPointMetrics(deviceGiD definition.DeviceGidType, points []model2.DataPoint) {
	go func() {
		deviceInfo, ok := cm.Worker().GetDeviceInfo(deviceGiD)
		if !ok {
			return
		}
		dimension := getAttrList(
			map[string]string{
				consts.AttrChannel:    deviceInfo.ChannelID,
				consts.AttrTemplate:   deviceInfo.Template,
				consts.AttrAContainer: utils.GetHostName(),
				consts.AttrDeviceName: deviceInfo.Name,
				consts.AttrDeviceGid:  string(deviceInfo.Gid),
			})

		m := make([]*metrics.Metrics, 0, len(points))
		for i := range points {
			_, pointAttrID, err := definition.SplitDataPointID(points[i].ID)
			if err != nil {
				continue
			}
			value, _ := points[i].Rtd.Val.Pv.AsDouble()
			m = append(m, metrics.NewMetrics(string(pointAttrID), value, metrics.PolicySET))
		}
		if len(m) == 0 {
			return
		}
		monitor.ReportMultiMetricsWithDimensions(m, dimension)
	}()
}
