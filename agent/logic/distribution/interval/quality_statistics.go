package interval

import (
	"fmt"
	"agent/entity/consts"
	"agent/logic/collector/device/virtualpoints"

	"trpc.group/trpc-go/trpc-go/log"

	"agent/entity/definition"
	"agent/logic/collector/rtdb"
	"agent/logic/collector/rtdb/model"
)

const LogShowIDCount = 60

var tmsDelayIDDict = map[int]int{
	definition.HalfMinutesTmsDelay: 0,
	definition.OneMinutesTmsDelay:  1,
}

type errorPoint struct {
	count int
	IDs   []definition.DataPointIDType
}

// AddIdToLogShow 添加测点ID到日志展示
func (e *errorPoint) AddIdToLogShow(id definition.DataPointIDType) {
	if len(e.IDs) > LogShowIDCount {
		return
	}
	e.IDs = append(e.IDs, id)
}

type qualityStatistics struct {
	deviceGid                 definition.DeviceGidType
	tmsDelayInfo              map[int]*errorPoint
	existTmsDelay             bool
	originalQuaErrorInfo      errorPoint //  原始质量标签异常的测点统计
	afterDeNoisedQuaErrorInfo errorPoint //  降噪后的质量标签异常的测点统计
}

// NewQualityStatistics 创建质量统计
func NewQualityStatistics(deviceGiD definition.DeviceGidType) *qualityStatistics {
	info := make(map[int]*errorPoint)
	for k, _ := range tmsDelayIDDict {
		info[k] = &errorPoint{}
	}
	return &qualityStatistics{
		deviceGid:    deviceGiD,
		tmsDelayInfo: info,
	}
}

// CountTmsDelay 计算时延
func (q *qualityStatistics) CountTmsDelay(pointID definition.DataPointIDType, delayTms int64) {
	var info *errorPoint
	if delayTms >= definition.OneMinutesTmsDelay {
		info = q.tmsDelayInfo[definition.OneMinutesTmsDelay]
	} else if delayTms >= definition.HalfMinutesTmsDelay {
		info = q.tmsDelayInfo[definition.HalfMinutesTmsDelay]
	}

	if info != nil {
		q.existTmsDelay = true
		info.count++
		info.AddIdToLogShow(pointID)
	}
}

// CountQuaError 计算质量标签异常
func (q *qualityStatistics) CountQuaError(pointID definition.DataPointIDType,
	originalQua consts.Quality, afterDeNoisedQua consts.Quality) {
	if originalQua != consts.QualityOk {
		q.originalQuaErrorInfo.count++
		q.originalQuaErrorInfo.AddIdToLogShow(pointID)
	}
	if afterDeNoisedQua != consts.QualityOk {
		q.afterDeNoisedQuaErrorInfo.count++
		q.afterDeNoisedQuaErrorInfo.AddIdToLogShow(pointID)
	}
}

// Report 上报
func (q *qualityStatistics) Report() {
	if q.existTmsDelay {
		logStr := fmt.Sprintf("TimestampStat ")
		for k, v := range q.tmsDelayInfo {
			if v.count > 0 {
				logStr += fmt.Sprintf("Delay:%v, Count:%v ,IDs:%v...;", k, v.count, v.IDs)
			}
		}
		log.Info(logStr)
	}
	if q.originalQuaErrorInfo.count > 0 {
		logStr := fmt.Sprintf("QualityStat original qua error count:%v", q.originalQuaErrorInfo.count)
		if q.afterDeNoisedQuaErrorInfo.count > 0 {
			logStr += fmt.Sprintf(", After denoise qua error count:%v", q.afterDeNoisedQuaErrorInfo.count)
		}
		log.Debug(logStr)
	}

	// 写实时数据
	points := make([]model.DataPoint, len(q.tmsDelayInfo))
	i := 0
	for k, v := range q.tmsDelayInfo {
		points[i] = model.NewVirtualDataPoint(q.deviceGid, definition.GeneratePointTmsDelayCountID(q.deviceGid, k))
		points[i].SetValue(v.count)
		i++
	}
	pointQua := []model.DataPoint{
		model.NewVirtualDataPointWithValue(q.deviceGid, definition.GeneratePointOriginQuaErrorCountID(q.deviceGid),
			q.originalQuaErrorInfo.count),
		model.NewVirtualDataPointWithValue(q.deviceGid, definition.GeneratePointQuaErrorCountID(q.deviceGid),
			q.afterDeNoisedQuaErrorInfo.count),
	}
	points = append(points, pointQua...)
	rtdb.SetDataPoints(points, false)

	// 上报监控
	virtualpoints.ReportDataPointMetrics(q.deviceGid, points)

}
