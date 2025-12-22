// Package read  通过store服务提供的API查询数据
package read

import (
	"context"
	"data-query/entity"
	conf "data-query/entity/config"
	"data-query/utils"
	"etrpc-go/log"
	"sort"
	"sync"
	"time"
	"trpcprotocol/data-cache"
)

func init() {
	Register("cache-api", &CacheApiRead{})
}

// CacheApiRead 缓存读取
type CacheApiRead struct {
	ReadMinutes  int                         `yaml:"read_minutes"`
	ReadDuration time.Duration               `yaml:"-"`
	cli          data_cache.PointClientProxy `yaml:"-"`
}

// GetType 判断插件类型
func (s CacheApiRead) GetType() string {
	return entity.ReadCacheType
}

// Setup 构建缓存查询插件
func (s CacheApiRead) Setup(cfg PlgConfig) (IReadPlugin, error) {
	plugin := &CacheApiRead{
		ReadMinutes: 10,
		cli:         data_cache.NewPointClientProxy(),
	}
	if err := cfg.Extra.Decode(plugin); err != nil {
		return nil, err
	}
	plugin.ReadDuration = time.Minute * time.Duration(plugin.ReadMinutes)
	return plugin, nil
}

// CanRead 判断当前时间范围能否查询
func (s CacheApiRead) CanRead(begin, end int64) bool {
	if begin == 0 {
		// begin=0,读取最新的值
		return true
	}
	// 查询的时间必须在本地缓存的时间范围内（引入余量处理临界情况，短时间过期依然允许查缓存）
	now := time.Now()
	minTs := now.Add(-(s.ReadDuration + time.Duration(conf.ServerConf.ExpireTimeMargin)*time.Second)).Unix()
	return begin >= minTs && end >= minTs
}

// ReadRange 查询时间范围数据
func (s CacheApiRead) ReadRange(ctx context.Context, pointNames []string, begin, end int64) (map[string][]*entity.Point, error) {
	req := &data_cache.ReqQueryData{
		PointList: pointNames,
		Begin:     begin,
		End:       end,
	}
	//Start1 := time.Now()
	rsp, err := s.cli.QueryData(ctx, req)
	if err != nil {
		return nil, err
	}
	//elapsedApi := int(time.Since(Start1).Milliseconds())
	//if elapsedApi > conf.ServerConf.NormalCostThreshold {
	//	log.Infof("存储服务调用, rpc查询耗时:%dms，req:%v,rsp:%v", elapsedApi, req, rsp)
	//}
	resMap := make(map[string][]*entity.Point)

	// 将接口结果转换为返回结果
	for pointName, pointValueList := range rsp.PointMap {
		var points []*entity.Point
		for _, p := range pointValueList.Values {
			points = append(points, &entity.Point{
				Name:  pointName,
				Value: p.V,
				Time:  p.T,
			})
		}
		// 按照timestamp从小到大排序
		sort.Slice(points, func(i, j int) bool {
			return points[i].Time < points[j].Time
		})
		resMap[pointName] = points
	}

	return resMap, nil
}

// ReadLatest 最近数据读取
func (s CacheApiRead) ReadLatest(ctx context.Context, pointName []string, max int64) (map[string]*entity.Point, error) {
	//TODO implement me
	panic("implement me")
}

// ReadChanged 变化数据读取
func (s CacheApiRead) ReadChanged(ctx context.Context, pointName []string, begin int64, end int64) (map[string]int64, error) {
	batchSize := conf.ServerConf.QueryChangedBatchSize
	if batchSize == 0 {
		batchSize = entity.DefaultQueryChangedBatchSize
	}
	batchList, err := utils.GetBatchStringList(pointName, batchSize)
	if err != nil {
		return nil, err
	}
	var retList = make([]map[string]int64, len(batchList))
	var localWg sync.WaitGroup
	maxConcurrency := conf.ServerConf.QueryChangedConcurrencyLimit
	if maxConcurrency == 0 {
		maxConcurrency = entity.DefaultQueryChangedConcurrencyLimit
	}
	semaphore := make(chan struct{}, maxConcurrency)

	for index, batch := range batchList {
		localWg.Add(1)
		go func(pos int, item []string) {
			defer localWg.Done()
			// 获取信号量
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			req := &data_cache.ReqQueryChanged{
				PointList: item,
			}
			if begin > 0 {
				req.Begin = begin
			}
			// 查询缓存服务
			batchRsp, err := s.cli.QueryChanged(ctx, req)
			if err != nil {
				log.Errorf("QueryChanged rpc req err:%s", err.Error())
				retList[pos] = make(map[string]int64)
			} else {
				retList[pos] = batchRsp.ChangedMap
			}
		}(index, batch)
	}
	localWg.Wait()
	// 合并所有分片的结果
	ret := make(map[string]int64)
	for _, v := range retList {
		for k, val := range v {
			ret[k] = val
		}
	}
	return ret, nil
}
