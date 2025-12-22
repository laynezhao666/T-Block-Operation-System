package point

import (
	"fmt"
	"math"
	"time"

	"alarm-compute/conf"
	"alarm-compute/utils/tnql"
)

// GetDurationInterval 根据时间和测点数量获取查询测点数据的间隔，适用于获取一段时间的测点数据
// duration
func GetDurationInterval(duration int64) (interval int, err error) {
	if duration < 0 {
		err = fmt.Errorf("bad duration: %v", duration)
		return
	}

	durationInt := int(duration)
	if durationInt > tnql.SecondsInDay {
		// 超过一天，最大取 1440 个点，调整 interval，10000个数据 = 1440 * 6 个点
		interval = int(math.Ceil(float64(duration) / float64(tnql.MaxOnePointNum)))
		return
	}

	// 超过一小时，10min取一个点
	if durationInt > 60*tnql.SecondsInMinute {
		interval = tnql.TenMinuteInterval
		return
	}

	if durationInt > 10*tnql.SecondsInMinute {
		// 超过10分钟，60秒取一个点
		interval = tnql.MinuteInterval
		return
	}

	if durationInt > tnql.SecondsInMinute {
		// 超过1分钟，5秒取一个点，10分钟 120 个点
		interval = tnql.SecondInterval
		return
	}

	interval = tnql.DefaultInterval
	return
}

// GetDataPointFetchNum 单次请求 数据模块 的测点数量限制
func GetDataPointFetchNum(interval int, start, end time.Time) (num int, err error) {
	// 单次请求的测点数量 = 单次请求能获取的最大数量 / 单个测点获取的数量
	// 单次请求能获取的最大数量 = MaxHBasePointFetchNum
	// 单个测点获取的数量 = end.Sub(start).Seconds() / interval

	if interval == 0 {
		// bad params
		err = fmt.Errorf("bad interval: %v", interval)
		return
	}

	duration := int(end.Sub(start).Seconds())
	if duration < 0 {
		err = fmt.Errorf("bad duration: %v, start: %v, end: %v", duration, start, end)
		return
	}
	if duration == 0 {
		num = int(conf.ServerConf.MaxPointCountForDataQuery)
		return
	}

	num = (int(conf.ServerConf.MaxPointCountForDataQuery) * interval) / duration
	return
}
