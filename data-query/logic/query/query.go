// Package query provides basic query functions.
package query

import (
	"context"
	"data-query/entity"
	conf "data-query/entity/config"
	"data-query/repo/read"
	"etrpc-go/log"
	"sort"
	"time"
	"trpcprotocol/data-query"

	"github.com/pkg/errors"
)

// DataQueryHandler 数据查询处理函数
func DataQueryHandler(ctx context.Context, req *data_query.QueryRequest) (*data_query.QueryResponse, error) {
	initStart := time.Now()
	// 1. 初始化
	resultData := make(map[string]*data_query.InnerMap)
	missPoints := make([]string, 0)

	// 2. 获取查询时间区间，间隔
	beginTime := req.Begin
	endTime := req.End
	interval := req.Interval
	if interval == 0 {
		interval = 1
	}

	// 数量限制和时间区间检查
	startTime := time.Unix(0, beginTime)
	endTimeObj := time.Unix(0, endTime)
	timeDuration := endTimeObj.Sub(startTime)
	if interval < conf.ServerConf.IntervalLimit && timeDuration >= 31*time.Minute {
		return nil, errors.New("查询时间区间不能超过31分钟，当间隔小于60分钟时")
	}

	// 3. 通过读取插件来查询实际数据
	res, readType, err := read.BatchReadRangePoints(ctx, req.PointList, beginTime, endTime)
	readCost := int(time.Since(initStart).Milliseconds())
	if err != nil {
		log.Error(err.Error())
		// 查不到不报错 (10.21sanshui沟通)
		return &data_query.QueryResponse{GetPointData: nil, MissPointName: req.PointList}, nil
	}

	pointProcessStart := time.Now()

	for _, pointName := range req.PointList {
		if !ContainsKey(res, pointName) {
			missPoints = append(missPoints, pointName)
			continue
		}
		var inner map[int64]float64
		if beginTime == 0 {
			// 只返回一个最新值
			inner = processLatestPoint(res[pointName])
		} else if beginTime == endTime {
			// 单时间点数据处理
			inner = processSinglePoint(res[pointName], beginTime)
		} else if beginTime < endTime {
			// 时间段数据处理
			inner = processRangePoint(res[pointName], beginTime, endTime, interval)
		}
		if len(inner) > 0 {
			resultData[pointName] = &data_query.InnerMap{InnerMap: inner}
		} else {
			missPoints = append(missPoints, pointName)
		}
	}
	pointProcessCost := int(time.Since(pointProcessStart).Milliseconds())
	totalCost := int(time.Since(initStart).Milliseconds())
	if totalCost > conf.ServerConf.ExtremelyCostThreshold {
		log.Infof("数据查询总耗时:%dms, 调用下游耗时:%dms, 调用类型:%s, 测点填补耗时:%dms, "+
			"begin:%d, end:%d, interval:%d, 请求测点:%v",
			totalCost, readCost, readType, pointProcessCost, beginTime, endTime, interval, req.PointList)
	}
	return &data_query.QueryResponse{GetPointData: resultData, MissPointName: missPoints}, nil
}

// 按时间点对数据进行填充
func processSinglePoint(datas []*entity.Point, queryTime int64) map[int64]float64 {
	inner := make(map[int64]float64)
	// 找到不小于目标时间点的最近的数据点
	index := sort.Search(len(datas), func(i int) bool {
		return datas[i].Time >= queryTime
	})
	if index < len(datas) && datas[index].Time == queryTime {
		// 时间完全匹配
		inner[queryTime] = datas[index].Value
	} else if index > 0 && isValidPoint(datas[index-1], queryTime) {
		// 否则用前一个数据点填补
		inner[queryTime] = datas[index-1].Value
	}
	return inner
}

// 按时间段对数据进行填充
func processRangePoint(datas []*entity.Point, beginTime, endTime int64, interval int64) map[int64]float64 {
	inner := make(map[int64]float64)
	n := len(datas)
	// 数据已经排序，双指针处理，时间复杂度:O(n * T/I)
	dataIndex := 0
	for queryTime := beginTime; queryTime <= endTime; queryTime += interval {
		// 移动 dataIndex 直到找到大于等于 queryTime 的数据点
		for dataIndex < n && datas[dataIndex].Time < queryTime {
			dataIndex++
		}
		if dataIndex < n && datas[dataIndex].Time == queryTime {
			inner[queryTime] = datas[dataIndex].Value
		} else if dataIndex > 0 && isValidPoint(datas[dataIndex-1], queryTime) {
			inner[queryTime] = datas[dataIndex-1].Value
		}
	}
	return inner
}

// isValidPoint 检查指定测点的时间基于查询时间是否已经过期
func isValidPoint(point *entity.Point, queryTime int64) bool {
	// 过期时间范围
	expireDuration := time.Duration(conf.ServerConf.ExpireTimeSinceQuery) * time.Second
	expireTime := queryTime - expireDuration.Nanoseconds()/int64(time.Second)
	return point.Time > expireTime
}

// ContainsKey 检查 map 是否包含指定的键
func ContainsKey(m map[string][]*entity.Point, key string) bool {
	_, exists := m[key]
	return exists
}

// 最新值处理
func processLatestPoint(datas []*entity.Point) map[int64]float64 {
	if len(datas) != 1 {
		return nil
	}
	inner := make(map[int64]float64)
	inner[datas[0].Time] = datas[0].Value
	return inner
}
