package virtualpoints

import (
	utils2 "agent/utils"
	"math"
	"sync/atomic"
	"time"

	"trpc.group/trpc-go/trpc-go/log"

	"agent/entity/config"
	"agent/entity/consts"
	"agent/entity/definition"
	devicemodel "agent/logic/collector/device/model"
	"agent/logic/collector/rtdb"
	"agent/logic/collector/rtdb/model"

	"github.com/robfig/cron/v3"
	"trpc.group/trpc-go/trpc-go/metrics"
)

const (
	CommunicationNormalStr       = "正常"
	CommunicationAbnormalStr     = "质量异常"
	CommunicationInterruptionStr = "通讯中断"
)

var (
	maxAllowedFailedRequestCount        = 5
	maxChannelAllowedFailedRequestCount = maxAllowedFailedRequestCount - 1
	maxAllowedFailedRequestTimeDuration = 60 // 失败持续时间（秒）
)

// Init 初始化
func Init() {
	maxAllowedFailedRequestCount = config.LoadIntOrDefault(config.GetRB().Collector.Common.RequestFailedCount, 14)
	maxChannelAllowedFailedRequestCount = (maxAllowedFailedRequestCount * 3) >> 2
	if maxChannelAllowedFailedRequestCount < 1 {
		maxChannelAllowedFailedRequestCount = 1
	}
	maxAllowedFailedRequestTimeDuration = config.LoadIntOrDefault(config.GetRB().Collector.Common.RequestFailedTime, 60)
	if maxAllowedFailedRequestTimeDuration < 10 {
		maxAllowedFailedRequestTimeDuration = 10 // 最小为10秒
	}

}

// VirtualPoints 虚拟测点
type VirtualPoints struct {
	communicationStatusChangedTime time.Time
	lastCommunicationOkTime        time.Time

	onePeriodSuccessCollectPoint int     // 一个周期内采集成功的测点个数
	onePeriodPointThroughput     float32 // 一个周期内的测点吞吐量：采集成功测点个数 / 时间（秒）

	onePeriodRangeResponseTimeMs int64 // 一个周期内响应时间极差
	onePeriodMaxResponseTimeMs   int64 // 一个周期内最大响应时间
	onePeriodMinResponseTimeMs   int64 // 一个周期内最小响应时间
	onePeriodCostTimeMs          int64 // 一个周期完成所花时间(毫秒)

	onePeriodSuccessRequestCount int // 一个周期内请求的成功次数
	onePeriodTotalRequestCount   int // 一个周期内请求的总次数

	successRequestCount atomic.Uint64 // 该设备请求的成功次数
	totalRequestCount   atomic.Uint64 // 该设备请求的总次数
	failedRequestCount  int           // 该设备请求的连续失败次数

	timeoutRequestCount uint64 // 该设备超时的次数

	totalRequestMessageCount   uint64 // 该设备与设备通讯报文总数
	successRequestMessageCount uint64 // 该设备与设备通讯报文成功数

	communicationStatusPoint          model.DataPoint
	pointThroughputPoint              model.DataPoint
	rangeResponseTimePoint            model.DataPoint
	maxResponseTimePoint              model.DataPoint
	minResponseTimePoint              model.DataPoint
	totalResponseTimePoint            model.DataPoint
	avgResponseTimePoint              model.DataPoint
	onePeriodSuccessRequestCountPoint model.DataPoint
	onePeriodTotalRequestCountPoint   model.DataPoint
	successRequestCountPoint          model.DataPoint
	minuteSuccessRequestCountPoint    model.DataPoint
	totalRequestCountPoint            model.DataPoint
	minuteTotalRequestCountPoint      model.DataPoint
	interruptionPoint                 model.DataPoint
	timeoutRequestCountPoint          model.DataPoint
	totalRequestMessageCountPoint     model.DataPoint
	successRequestMessageCountPoint   model.DataPoint

	// 业务相关测点
	commPoint         model.DataPoint
	channelCommPoints []model.DataPoint // 每个通道的通讯状态
	channelNumber     int

	// 上报属性
	deviceAttrList []*metrics.Dimension
	// 各通道的上报属性
	channelAttrList [][]*metrics.Dimension

	c                        *cron.Cron
	firstCalculateMinuteData bool
	lastSuccessCount         uint64
	lastTotalCount           uint64
	currentSuccessCount      uint64
	currentTotalCount        uint64
}

// NewVirtualPoints 初始化虚拟测点
func NewVirtualPoints(deviceGiD definition.DeviceGidType,
	channels []devicemodel.Channel) *VirtualPoints {
	v := &VirtualPoints{
		pointThroughputPoint:              model.NewVirtualDataPoint(deviceGiD, definition.GeneratePointThroughputID(deviceGiD)),
		communicationStatusPoint:          model.NewVirtualDataPoint(deviceGiD, definition.GenerateCommunicationStatusID(deviceGiD)),
		rangeResponseTimePoint:            model.NewVirtualDataPoint(deviceGiD, definition.GenerateRangeResponseTimeID(deviceGiD)),
		maxResponseTimePoint:              model.NewVirtualDataPoint(deviceGiD, definition.GenerateMaxResponseTimeID(deviceGiD)),
		minResponseTimePoint:              model.NewVirtualDataPoint(deviceGiD, definition.GenerateMinResponseTimeID(deviceGiD)),
		totalResponseTimePoint:            model.NewVirtualDataPoint(deviceGiD, definition.GenerateTotalResponseTimeID(deviceGiD)),
		avgResponseTimePoint:              model.NewVirtualDataPoint(deviceGiD, definition.GenerateAvgResponseTimeID(deviceGiD)),
		onePeriodSuccessRequestCountPoint: model.NewVirtualDataPoint(deviceGiD, definition.GenerateOnePeriodSuccessRequestCountID(deviceGiD)),
		onePeriodTotalRequestCountPoint:   model.NewVirtualDataPoint(deviceGiD, definition.GenerateOnePeriodTotalRequestCountID(deviceGiD)),
		successRequestCountPoint:          model.NewVirtualDataPoint(deviceGiD, definition.GenerateSuccessRequestCountID(deviceGiD)),
		minuteSuccessRequestCountPoint:    model.NewVirtualDataPoint(deviceGiD, definition.GenerateMinuteSuccessRequestCountID(deviceGiD)),
		totalRequestCountPoint:            model.NewVirtualDataPoint(deviceGiD, definition.GenerateTotalRequestCountID(deviceGiD)),
		minuteTotalRequestCountPoint:      model.NewVirtualDataPoint(deviceGiD, definition.GenerateMinuteTotalRequestCountID(deviceGiD)),
		interruptionPoint:                 model.NewVirtualDataPoint(deviceGiD, definition.GenerateInterruptionID(deviceGiD)),
		timeoutRequestCountPoint:          model.NewVirtualDataPoint(deviceGiD, definition.GenerateTimeoutRequestCountID(deviceGiD)),
		totalRequestMessageCountPoint:     model.NewVirtualDataPoint(deviceGiD, definition.GenerateTotalRequestMessageCountID(deviceGiD)),
		successRequestMessageCountPoint:   model.NewVirtualDataPoint(deviceGiD, definition.GenerateSuccessRequestMessageCountID(deviceGiD)),

		// 业务相关测点
		commPoint: model.NewVirtualDataPoint(deviceGiD, definition.GenerateCommID(deviceGiD)),
	}
	channelNumber := len(channels)
	v.channelCommPoints = make([]model.DataPoint, 0, channelNumber)
	v.channelNumber = channelNumber
	for i := 0; i < channelNumber; i++ {
		v.channelCommPoints = append(v.channelCommPoints,
			model.NewVirtualDataPoint(deviceGiD, definition.GenerateChannelCommID(deviceGiD, i)))
	}

	v.c = cron.New(cron.WithSeconds())
	_, _ = v.c.AddFunc("0 * * * * *", func() {
		v.calculateMinuteData()
	})
	v.start()

	v.Clear()

	v.lastCommunicationOkTime = utils2.GetNowUTCTime()

	return v
}

func (v *VirtualPoints) start() {
	v.firstCalculateMinuteData = true
	v.c.Start()
}

// ResetValueAfterOnePeriod 重置周期内的变量
func (v *VirtualPoints) ResetValueAfterOnePeriod() {
	if v == nil {
		return
	}

	v.onePeriodSuccessCollectPoint = 0

	v.onePeriodPointThroughput = 0.0

	v.onePeriodRangeResponseTimeMs = 0
	v.onePeriodMaxResponseTimeMs = math.MinInt64
	v.onePeriodMinResponseTimeMs = math.MaxInt64
	v.onePeriodCostTimeMs = 0

	v.onePeriodSuccessRequestCount = 0
	v.onePeriodTotalRequestCount = 0
}

// Clear 清空数据
func (v *VirtualPoints) Clear() {
	if v == nil {
		return
	}

	// 跨周期的变量
	v.failedRequestCount = 0
	v.totalRequestMessageCount = 0
	v.successRequestMessageCount = 0

	v.communicationStatusChangedTime = time.Unix(0, 0)

	// 周期内的变量
	v.ResetValueAfterOnePeriod()
}

// AddPeriodCostTime 增加周期内的耗时
func (v *VirtualPoints) AddPeriodCostTime(costTime int64) {
	if v == nil {
		return
	}

	if costTime > v.onePeriodMaxResponseTimeMs {
		v.onePeriodMaxResponseTimeMs = costTime
	}

	if costTime < v.onePeriodMinResponseTimeMs {
		v.onePeriodMinResponseTimeMs = costTime
	}

	v.onePeriodCostTimeMs += costTime
}

// UpdateAfterOnePeriodFinished 更新周期内的数据
func (v *VirtualPoints) UpdateAfterOnePeriodFinished(packetNum int) {
	if v == nil {
		return
	}

	v.updatePointThroughput()
	v.updateResponseTime(packetNum)
}

// UpdateAfterOneRequestFinished 返回是否中断
func (v *VirtualPoints) UpdateAfterOneRequestFinished(currentRequestSuccess bool, pointNum int,
	totalMessageCount uint64, successMessageCount uint64) bool {
	if v == nil {
		return false
	}

	v.updateRequestCount(currentRequestSuccess, pointNum, totalMessageCount, successMessageCount)
	return v.updateCommStatus()
}

// AddAndUpdateTimeoutNumber 增加超时数量
func (v *VirtualPoints) AddAndUpdateTimeoutNumber(isTimeout bool) {
	if v == nil {
		return
	}

	if isTimeout {
		v.timeoutRequestCount++
		v.reportTimeoutRequest()
	}

	v.timeoutRequestCountPoint.SetValue(v.timeoutRequestCount)
	rtdb.SetDataPoints(model.DataPoints{v.timeoutRequestCountPoint}, false)
}

func (v *VirtualPoints) updateRequestCount(currentRequestSuccess bool, pointNum int,
	totalMessageCount uint64, successMessageCount uint64) {
	v.totalRequestCount.Add(1)
	v.onePeriodTotalRequestCount++
	v.totalRequestMessageCount += totalMessageCount
	v.successRequestMessageCount += successMessageCount

	if currentRequestSuccess {
		v.successRequestCount.Add(1)
		v.onePeriodSuccessRequestCount++

		v.onePeriodSuccessCollectPoint += pointNum

		v.failedRequestCount = 0
	} else {
		v.failedRequestCount++
	}

	// 为避免上报后 成功请求数 > 总请求数，同时上报
	v.reportRequestCount(currentRequestSuccess, v.totalRequestMessageCount, v.successRequestMessageCount)

	currentTime := utils2.GetNowUTCTimeStamp()
	v.onePeriodSuccessRequestCountPoint.SetValueWithTime(v.onePeriodSuccessRequestCount, currentTime)
	v.onePeriodTotalRequestCountPoint.SetValueWithTime(v.onePeriodTotalRequestCount, currentTime)
	v.successRequestCountPoint.SetValueWithTime(v.successRequestCount.Load(), currentTime)
	v.totalRequestCountPoint.SetValueWithTime(v.totalRequestCount.Load(), currentTime)
	v.successRequestMessageCountPoint.SetValueWithTime(v.successRequestMessageCount, currentTime)
	v.totalRequestMessageCountPoint.SetValueWithTime(v.totalRequestMessageCount, currentTime)
	rtdb.SetDataPoints([]model.DataPoint{
		v.onePeriodSuccessRequestCountPoint,
		v.onePeriodTotalRequestCountPoint,
		v.successRequestCountPoint,
		v.totalRequestCountPoint,
		v.successRequestMessageCountPoint,
		v.totalRequestMessageCountPoint},
		false)
}

func (v *VirtualPoints) updatePointThroughput() {
	t := v.onePeriodCostTimeMs
	if t <= 0 {
		t = 1
	}

	v.onePeriodPointThroughput = 1000 * float32(v.onePeriodSuccessCollectPoint) / float32(t)
	v.pointThroughputPoint.SetValue(v.onePeriodPointThroughput)
	rtdb.SetDataPoints([]model.DataPoint{v.pointThroughputPoint}, false)

	v.reportPointThroughput()
}

func (v *VirtualPoints) updateResponseTime(packetNum int) {
	avgTime := v.onePeriodCostTimeMs
	if packetNum > 0 {
		avgTime /= int64(packetNum)
	}
	v.onePeriodRangeResponseTimeMs = v.onePeriodMaxResponseTimeMs - v.onePeriodMinResponseTimeMs
	if v.onePeriodRangeResponseTimeMs < 0 {
		v.onePeriodRangeResponseTimeMs = 0
	}

	currentTime := utils2.GetNowUTCTimeStamp()
	v.avgResponseTimePoint.SetValueWithTime(avgTime, currentTime)
	v.totalResponseTimePoint.SetValueWithTime(v.onePeriodCostTimeMs, currentTime)
	v.minResponseTimePoint.SetValueWithTime(v.onePeriodMinResponseTimeMs, currentTime)
	v.maxResponseTimePoint.SetValueWithTime(v.onePeriodMaxResponseTimeMs, currentTime)
	v.rangeResponseTimePoint.SetValueWithTime(v.onePeriodRangeResponseTimeMs, currentTime)
	rtdb.SetDataPoints(model.DataPoints{
		v.avgResponseTimePoint,
		v.totalResponseTimePoint,
		v.minResponseTimePoint,
		v.maxResponseTimePoint,
		v.rangeResponseTimePoint,
	}, false)

	v.reportPeriodResponseTime()
}

func (v *VirtualPoints) isCommunicationInterruption() bool {
	countInterruption := IsDeviceCommunicationInterruption(v.failedRequestCount)
	timeInterruption := IsCommunicationInterruptionByTimeDuration(
		v.lastCommunicationOkTime, maxAllowedFailedRequestTimeDuration)
	log.Debugf("failCount:%v, countInterruption:%v, timeInterruption:%v", v.failedRequestCount, countInterruption,
		timeInterruption)
	return countInterruption || timeInterruption
}

// GetChannelInterruptionStatus 获取索引为 index 的通道的通讯状态
func (v *VirtualPoints) GetChannelInterruptionStatus(index int) bool {
	if index < 0 || index >= v.channelNumber {
		// 如果越界，视为中断
		return true
	}
	b, err := v.channelCommPoints[index].Rtd.Val.Pv.AsInt64()
	if err != nil {
		// 数值转换错误，视为中断
		return true
	}
	return b >= 1
}

// UpdateChannelInterruptionStatus 更新索引为 index 的通道的通讯状态
func (v *VirtualPoints) UpdateChannelInterruptionStatus(isInterrupted bool, index int, changedTimestamp int64) {
	if index < 0 || index >= v.channelNumber {
		return
	}

	v.channelCommPoints[index].SetValueWithTime(utils2.Bool2Int(isInterrupted), changedTimestamp)
	rtdb.SetDataPoints([]model.DataPoint{v.channelCommPoints[index]}, true)
}

func (v *VirtualPoints) updateInterruptionStatus(isInterrupted int) {
	t := v.communicationStatusChangedTime.Unix()
	// 程序内部使用 v.interruptionPoint 作为通讯状态
	v.interruptionPoint.SetValueWithTime(isInterrupted, t)
	v.commPoint.SetValueWithTime(isInterrupted, t)

	rtdb.SetDataPoints([]model.DataPoint{v.interruptionPoint, v.commPoint}, true)
}

// clearCommunicationInterruptionStatus 清空设备通讯中断状态
func (v *VirtualPoints) clearCommunicationInterruptionStatus(currentTime time.Time) {
	v.lastCommunicationOkTime = currentTime
	v.communicationStatusChangedTime = currentTime
	v.updateInterruptionStatus(0)
}

// 返回是否中断
func (v *VirtualPoints) updateCommStatus() bool {
	isInterrupted := false

	currentTime := utils2.GetNowUTCTime()
	currentTimestamp := currentTime.Unix()
	// 连续失败请求数 > 0
	if v.failedRequestCount > 0 {
		// 超过阈值，判定为通讯中断
		if v.isCommunicationInterruption() {
			v.communicationStatusChangedTime = currentTime
			isInterrupted = true
			v.communicationStatusPoint.SetValueWithTimeAndDesc(1, currentTimestamp, CommunicationInterruptionStr)
			v.updateInterruptionStatus(1)
		} else {
			v.communicationStatusPoint.SetValueWithTimeAndDesc(0, currentTimestamp, CommunicationAbnormalStr)
			v.communicationStatusPoint.Rtd.Val.Qua = consts.QualityCommAbnormal // 当重启时，避免上次的通讯中断告警恢复
		}
	} else {
		v.clearCommunicationInterruptionStatus(currentTime)
		v.communicationStatusPoint.SetValueWithTimeAndDesc(0, currentTimestamp, CommunicationNormalStr)
	}

	rtdb.SetDataPoints(model.DataPoints{v.communicationStatusPoint}, true)

	return isInterrupted
}

// GetPoints 获取虚拟测点
func (v *VirtualPoints) GetPoints() definition.DataPointIDsType {
	if v == nil {
		return nil
	}

	r := definition.DataPointIDsType{
		v.pointThroughputPoint.ID,
		v.communicationStatusPoint.ID,
		v.rangeResponseTimePoint.ID,
		v.maxResponseTimePoint.ID,
		v.minResponseTimePoint.ID,
		v.totalResponseTimePoint.ID,
		v.avgResponseTimePoint.ID,
		v.onePeriodSuccessRequestCountPoint.ID,
		v.onePeriodTotalRequestCountPoint.ID,
		v.successRequestCountPoint.ID,
		v.totalRequestCountPoint.ID,
		v.interruptionPoint.ID,
		v.timeoutRequestCountPoint.ID,
		v.commPoint.ID,
	}
	for i := range v.channelCommPoints {
		r = append(r, v.channelCommPoints[i].ID)
	}
	return r
}

func (v *VirtualPoints) calculateMinuteData() {
	if v == nil {
		return
	}

	for {
		v.currentSuccessCount = v.successRequestCount.Load()
		v.currentTotalCount = v.totalRequestCount.Load()
		if v.firstCalculateMinuteData {
			break
		}

		minuteSuccessReqCount := utils2.SubtractUint64(v.currentSuccessCount, v.lastSuccessCount)
		minuteTotalRequestCount := utils2.SubtractUint64(v.currentTotalCount, v.lastTotalCount)

		currentTime := utils2.GetNowUTCTimeStamp()

		v.minuteSuccessRequestCountPoint.SetValueWithTime(minuteSuccessReqCount, currentTime)
		v.minuteTotalRequestCountPoint.SetValueWithTime(minuteTotalRequestCount, currentTime)

		rtdb.SetDataPoints(model.DataPoints{
			v.minuteSuccessRequestCountPoint,
			v.minuteTotalRequestCountPoint,
		}, false)
		break
	}

	v.firstCalculateMinuteData = false
	v.lastSuccessCount = v.currentSuccessCount
	v.lastTotalCount = v.currentTotalCount
}

// Close 关闭虚拟点
func (v *VirtualPoints) Close() {
	v.c.Stop()
}
