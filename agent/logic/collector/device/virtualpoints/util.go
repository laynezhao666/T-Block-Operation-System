package virtualpoints

import (
	"agent/utils"
	"time"

	"trpc.group/trpc-go/trpc-go/metrics"
)

// IsDeviceCommunicationInterruption 设备通讯中断
func IsDeviceCommunicationInterruption(failedCount int) bool {
	return failedCount >= maxAllowedFailedRequestCount
}

// IsChannelCommunicationInterruption 通道通讯中断
func IsChannelCommunicationInterruption(failedCount int) bool {
	return failedCount >= maxChannelAllowedFailedRequestCount
}

// IsCommunicationInterruptionByTimeDuration 通讯中断
func IsCommunicationInterruptionByTimeDuration(lastOkTime time.Time, allowTimeoutSecond int) bool {
	return utils.GetNowUTCTimeStamp()-lastOkTime.Unix() > int64(allowTimeoutSecond)
}

func getAttrList(attrs map[string]string) []*metrics.Dimension {
	attrsList := make([]*metrics.Dimension, 0, len(attrs))
	for n, v := range attrs {
		attrsList = append(attrsList, &metrics.Dimension{
			Name:  n,
			Value: v,
		})
	}
	return attrsList
}
