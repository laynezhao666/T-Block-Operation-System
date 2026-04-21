package virtualpoints

import (
	"dac/entity/config"
	"time"

	"dac/entity/consts"
	"dac/entity/model/db"
	"dac/entity/model/rt"
	"dac/entity/utils"
	"dac/logic/collect/rtdb"

	"dac/entity/utils/ttime"
	"trpc.group/trpc-go/trpc-go/metrics"
)

type VirtualPoints struct {
	communicationStatusChangedTime time.Time

	onePeriodCostTimeMs int64 // 一个周期完成所花时间(毫秒)

	successRequestCount uint64 // 该设备请求的成功次数
	totalRequestCount   uint64 // 该设备请求的总次数
	failedRequestCount  int    // 该设备请求的连续失败次数

	// 上次通讯状态：0=在线，1=离线，-1=未初始化
	lastCommStatus int64

	communicationStatusPoint rt.Point
	totalResponseTimePoint   rt.Point
	successRequestCountPoint rt.Point
	totalRequestCountPoint   rt.Point

	// 上报属性
	deviceAttrList []*metrics.Dimension
}

func NewVirtualPoints(deviceID db.IDType, deviceAttrs map[string]string) VirtualPoints {
	v := VirtualPoints{
		communicationStatusPoint: rt.NewPoint(utils.GenerateCommID(deviceID)),

		totalResponseTimePoint: rt.NewPoint(utils.GenerateTotalResponseTimeID(deviceID)),

		successRequestCountPoint: rt.NewPoint(utils.GenerateSuccessRequestCountID(deviceID)),
		totalRequestCountPoint:   rt.NewPoint(utils.GenerateTotalRequestCountID(deviceID)),
	}

	v.deviceAttrList = getAttrList(deviceAttrs)

	channelChannelAttrs := make(map[string]string, len(deviceAttrs))
	for k, vv := range deviceAttrs {
		channelChannelAttrs[k] = vv
	}

	v.Clear()
	return v
}

func (v *VirtualPoints) Clear() {
	if v == nil {
		return
	}
	// 跨周期的变量
	v.totalRequestCount = 0
	v.successRequestCount = 0
	v.failedRequestCount = 0
	v.lastCommStatus = -1 // 初始化为未知状态，确保首次更新会写入 Redis

	v.communicationStatusChangedTime = time.Unix(0, 0)

	// 周期内的变量
	v.ResetValueAfterOnePeriod()
}

func (v *VirtualPoints) GetCommStatusPoint() rt.Point {
	return v.communicationStatusPoint
}

func (v *VirtualPoints) ResetValueAfterOnePeriod() {
	if v == nil {
		return
	}

	v.onePeriodCostTimeMs = 0
}

func (v *VirtualPoints) AddPeriodCostTime(costTime int64) {
	if v == nil {
		return
	}

	v.onePeriodCostTimeMs += costTime
}

// UpdateAfterOneRequestFinished 返回是否中断
func (v *VirtualPoints) UpdateAfterOneRequestFinished(currentRequestSuccess bool, mozu string) bool {
	if v == nil {
		return false
	}

	v.updateRequestCount(currentRequestSuccess)
	return v.updateCommStatus(mozu)
}

func (v *VirtualPoints) UpdateAfterOnePeriodFinished() {
	if v == nil {
		return
	}

	v.updateResponseTime()
}

func (v *VirtualPoints) updateResponseTime() {
	currentTime := ttime.GetNowUTC().UnixMilli()
	v.totalResponseTimePoint.SetValueWithTime(v.onePeriodCostTimeMs, currentTime)
	rtdb.SetPoints(rt.Points{v.totalResponseTimePoint}, true)

	v.reportPeriodResponseTime()
}

func (v *VirtualPoints) updateRequestCount(currentRequestSuccess bool) {
	v.totalRequestCount++

	if currentRequestSuccess {
		v.successRequestCount++

		v.failedRequestCount = 0
	} else {
		v.failedRequestCount++
	}

	// 为避免上报后 成功请求数 > 总请求数，同时上报
	v.reportRequestCount(currentRequestSuccess)

	currentTime := ttime.GetNowUTC().UnixMilli()
	v.successRequestCountPoint.SetValueWithTime(v.successRequestCount, currentTime)
	v.totalRequestCountPoint.SetValueWithTime(v.totalRequestCount, currentTime)
	rtdb.SetPoints(rt.Points{
		v.successRequestCountPoint,
		v.totalRequestCountPoint},
		true)
}

// 返回是否中断
func (v *VirtualPoints) updateCommStatus(mozu string) bool {
	isInterrupted := false

	currentTime := ttime.GetNowUTC()
	currentTimestamp := currentTime.UnixMilli()

	// 计算当前通讯状态
	var newCommStatus int64 = 0 // 默认在线
	if v.failedRequestCount > 0 {
		if v.isCommunicationInterruption(mozu) {
			newCommStatus = 1 // 离线
			isInterrupted = true
		} else if v.successRequestCount == 0 {
			// 从未成功连接过的设备，不应默认标记为在线
			// 保持上次状态不变（如果是首次则不写入任何状态）
			return false
		}
	}

	// 只有状态变化时才更新 Redis（防抖优化）
	if v.lastCommStatus != newCommStatus {
		v.communicationStatusChangedTime = currentTime
		v.communicationStatusPoint.SetValueWithTime(newCommStatus, currentTimestamp)
		rtdb.SetPoints(rt.Points{v.communicationStatusPoint}, true)
		v.lastCommStatus = newCommStatus
	}

	return isInterrupted
}

func (v *VirtualPoints) isCommunicationInterruption(mozu string) bool {
	return IsDeviceCommunicationInterruption(v.failedRequestCount, mozu)
}

// IsDeviceCommunicationInterruption 设备通讯中断阈值判断
func IsDeviceCommunicationInterruption(failedCount int, mozu string) bool {
	maxCount := config.C.ToleranceMozuMaxCount(mozu)
	if maxCount > 0 {
		return failedCount >= maxCount
	}
	return failedCount >= maxAllowedFailedRequestCount
}

// clearCommInterruptStatus 清空设备通讯中断状态
func (v *VirtualPoints) clearCommInterruptStatus(currentTime time.Time) {
	v.communicationStatusChangedTime = currentTime
}

func (v *VirtualPoints) reportPeriodResponseTime() {
	v.reportMetric(consts.InternalIDTotalResponseTime, float64(v.onePeriodCostTimeMs), metrics.PolicySET)
}
