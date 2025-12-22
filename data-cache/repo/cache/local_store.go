// Package cache 本地数据缓存相关读写逻辑
package cache

import (
	"data-cache/entity/consts"
	"data-cache/entity/model"
	"data-cache/repo/report"
	"etrpc-go/config"
	"fmt"
	"time"
)

func init() {
	config.Register("local-store", storeCfg, config.WithPrefix("store"))
}

// localStoreCfg 本地存储相关配置信息
type localStoreCfg struct {
	BulkCnt     uint64        `yaml:"bulk_cnt"`     // 数据桶数量
	StoreMinute uint32        `yaml:"store_minute"` // 存储分钟数
	expire      time.Duration // 过期时间
}

var (
	storeCfg = &localStoreCfg{
		BulkCnt:     20000,
		StoreMinute: 11,
		expire:      11 * time.Minute,
	} // 本地存储配置
	DataCache *ShardMap[string, *PointCache] // 测点数据缓存
)

// Setup 初始化本地存储
func Setup() {
	SetWindowCfg(defaultWindowSize, storeCfg.StoreMinute)
	storeCfg.expire = time.Duration(storeCfg.StoreMinute) * time.Minute
	DataCache = NewShardMap[string, *PointCache](storeCfg.BulkCnt, DefaultHash)
}

// Write 写入本地缓存
func Write(points []*model.StdPoint) {
	begin := time.Now()
	expireTs := uint32(begin.Add(-storeCfg.expire).Unix())
	// 批量取出所有测点对应的缓存结构
	for _, point := range points {
		// 数据超过缓存时间范围
		if point.Time < expireTs {
			continue
		}
		// 获取测点缓存对象
		val, ok := DataCache.Get(point.Name)
		if !ok {
			// 首次设置，采用SetNx，避免多协程同时设置出现异常
			val = DataCache.SetNx(point.Name, NewPointCache())
		}
		// 写入测点数据
		val.Add(point.Time, &model.CachePoint{
			Quality: point.Quality,
			Value:   point.Value,
		})
		// 写入变化时间
		if point.Interval == consts.PointIntervalChanged {
			val.lastChangeTs = point.Time
		}
	}
	cost := time.Since(begin).Milliseconds()
	report.LocalWriteCnt.Report(float64(len(points)))
	report.LocalWriteCost.Report(float64(cost))
}

// CanRead 判断是否可以读取本地缓存，判断读取范围是否在时间范围内
func CanRead(begin, end int64) bool {
	minTime := time.Now().Add(-storeCfg.expire).Unix()
	return begin >= minTime && end >= minTime
}

// ReadRange 读取本地缓存
func ReadRange(pointNames []string, begin, end int64) (map[string][]*model.StdPoint, error) {
	if !CanRead(begin, end) {
		return nil, fmt.Errorf("read time range [%d,%d] over cache duration", begin, end)
	}
	beginTime := time.Now()
	res := map[string][]*model.StdPoint{}
	exist := DataCache.GetMany(pointNames)
	totalCnt := 0
	for pointName, val := range exist {
		data := val.Range(uint32(begin), uint32(end))
		pList := make([]*model.StdPoint, 0, len(data))
		for ts, v := range data {
			pList = append(pList, &model.StdPoint{
				Name:    pointName,
				Time:    uint32(ts),
				Value:   v.Value,
				Quality: v.Quality,
			})
		}
		totalCnt += len(pList)
		res[pointName] = pList
	}
	cost := time.Since(beginTime).Milliseconds()
	report.LocalReadCnt.Report(float64(totalCnt))
	report.LocalReadCost.Report(float64(cost))
	return res, nil
}

// ReadLatest 读取本地缓存某个时间点前最新数据
func ReadLatest(pointNames []string, max int64) (map[string]*model.StdPoint, error) {
	if !CanRead(max, max) {
		return nil, fmt.Errorf("read time [%d] over cache duration", max)
	}
	beginTime := time.Now()
	res := map[string]*model.StdPoint{}
	exist := DataCache.GetMany(pointNames)
	for pointName, val := range exist {
		ts, item := val.Latest(uint32(max))
		if item == nil {
			continue
		}
		res[pointName] = &model.StdPoint{
			Name:    pointName,
			Time:    ts,
			Value:   item.Value,
			Quality: item.Quality,
		}
	}
	cost := time.Since(beginTime).Milliseconds()
	report.LocalReadCnt.Report(float64(len(res)))
	report.LocalReadCost.Report(float64(cost))
	return res, nil
}

// ReadChanged 读取本地缓存某个时间点后变化的测点有哪些
func ReadChanged(pointNames []string, min int64) (map[string]int64, error) {
	beginTime := time.Now()
	exist := DataCache.GetMany(pointNames)
	res := make(map[string]int64)
	for pointName, val := range exist {
		ts := val.GetLastChangeTs()
		if ts >= uint32(min) {
			res[pointName] = int64(ts)
		}
	}
	cost := time.Since(beginTime).Milliseconds()
	report.LocalReadChangedCnt.Report(float64(len(res)))
	report.LocalReadChangedCost.Report(float64(cost))
	return res, nil
}
