package point

import (
	"context"
	"fmt"
	"sync"
	"time"

	"alarm-compute/conf"
	"alarm-compute/entity/epoint"
	"alarm-compute/repo"
	"alarm-compute/utils/common"
	"alarm-compute/utils/modcall"
	"etrpc-go/log"

	"github.com/panjf2000/ants/v2"
	"trpc.group/trpc-go/trpc-go"
)

var (
	pointManager *PointManager
	once         sync.Once
)

// PointManager 测点数据管理
type PointManager struct{}

type delayPoint struct {
	delay     int
	pointList []string
}

// GetPointManager 获取测点数据管理
func GetPointManager() *PointManager {
	once.Do(func() {
		pointManager = &PointManager{}
	})
	return pointManager
}

// BatchGetRTPointValue 批量查询测点实时数据
func (m *PointManager) BatchGetRTPointValue(ctx context.Context, points []string, t time.Time) (map[string]float64, error) {
	query := &repo.PointQueryReq{
		PointList: points,
		Begin:     t.Unix(),
		End:       t.Unix(),
		Interval:  1,
	}
	pointValue, err := repo.GetPointDataSvc().GetPointDataTS(trpc.BackgroundContext(), query)
	if err != nil {
		return nil, err
	}
	return pointValue, nil
}

// BatchGetIntervalPointValue 批量获取测点数据
func (m *PointManager) BatchGetIntervalPointValue(ctx context.Context, hPointMap map[int][]string, t time.Time,
	isVirtual bool) (epoint.HistoryValueMap, error) {
	var poolSize, batchSize int
	if isVirtual {
		poolSize = int(conf.ServerConf.VirtualConfig.IntervalRequestPoolSize)
		batchSize = int(conf.ServerConf.VirtualConfig.IntervalBatchPointSize)
	} else {
		poolSize = int(conf.ServerConf.DelayTimeConfig.IntervalRequestPoolSize)
		batchSize = int(conf.ServerConf.DelayTimeConfig.IntervalBatchPointSize)
	}
	dataCh := make(chan map[int]epoint.ValueMap, 2*poolSize)
	var getDataTask = func(i interface{}) error {
		pm := i.(delayPoint)
		historyData, err := m.GetIntervalPointValue(ctx, pm, t)
		if err != nil {
			log.Errorf("getIntervalValue Err %v", err)
			return err
		}
		dataCh <- map[int]epoint.ValueMap{pm.delay: historyData}
		return nil
	}
	ret := epoint.HistoryValueMap{}
	var dataHandleWg sync.WaitGroup
	dataHandleWg.Add(1)
	go func() {
		defer dataHandleWg.Done()
		for item := range dataCh {
			for delay, val := range item {
				for k, v := range val {
					if _, ok := ret[k]; !ok {
						ret[k] = make(map[int]float64)
					}
					ret[k][delay] = v
				}
			}
		}
	}()
	// getData
	var poolWg sync.WaitGroup
	wp, _ := ants.NewPoolWithFunc(poolSize, func(i interface{}) {
		getDataTask(i)
		poolWg.Done()
	})
	defer wp.Release()
	for k, v := range hPointMap {
		chunkList, err := common.ChunkStringList(v, batchSize)
		if err != nil {
			return nil, fmt.Errorf("ChunkList failed %w", err)
		}
		for _, item := range chunkList {
			i := delayPoint{
				delay:     k,
				pointList: item,
			}
			poolWg.Add(1)
			wp.Invoke(i)
		}
	}
	go func() {
		poolWg.Wait()
		close(dataCh)
	}()
	dataHandleWg.Wait()
	return ret, nil
}

// BatchGetDurationPointValue 获取一段时间测点的数据，函数需要用到
func (m *PointManager) BatchGetDurationPointValue(ctx context.Context,
	durationPM map[int][]string, t time.Time, isVirtual bool) (epoint.HistoryValueMap, error) {
	var poolSize int
	if isVirtual {
		poolSize = int(conf.ServerConf.VirtualConfig.DurationRequestPoolSize)
	} else {
		poolSize = int(conf.ServerConf.DelayTimeConfig.DurationRequestPoolSize)
	}
	dataCh := make(chan epoint.HistoryTimeValueMap, 2*poolSize)
	var getDataTask = func(i interface{}) error {
		pm := i.(delayPoint)
		historyData, err := m.GetDurationPointValue(ctx, pm, t)
		if err != nil {
			log.Errorf("getDurationValue Err %v", err)
			return err
		}
		dataCh <- historyData
		return nil
	}
	tUnix := t.Unix()
	ret := epoint.HistoryValueMap{}
	var dataHandleWg sync.WaitGroup
	dataHandleWg.Add(1)
	go func() {
		defer dataHandleWg.Done()
		for item := range dataCh {
			for p, valueMap := range item {
				if _, ok := ret[p]; !ok {
					ret[p] = epoint.IntervalMap{}
				}
				for ts, v := range valueMap {
					d := tUnix - ts
					ret[p][int(d)] = v
				}
			}
		}
	}()
	// getData
	var poolWg sync.WaitGroup
	wp, _ := ants.NewPoolWithFunc(poolSize, func(i interface{}) {
		getDataTask(i)
		poolWg.Done()
	})
	defer wp.Release()
	for duration, pointList := range durationPM {
		i := delayPoint{
			delay:     duration,
			pointList: pointList,
		}
		poolWg.Add(1)
		wp.Invoke(i)
	}
	go func() {
		poolWg.Wait()
		close(dataCh)
	}()
	dataHandleWg.Wait()
	return ret, nil
}

// BatchGetRangePointValue 获取一段时间跳变测点的数据，跳变函数专用
func (m *PointManager) BatchGetRangePointValue(ctx context.Context,
	rangePM map[int]map[int][]string, t time.Time, isVirtual bool) (epoint.HistoryValueMap, error) {

	data := epoint.HistoryValueMap{}
	for delay, pm := range rangePM {
		delayTime := t.Add(-1 * time.Second * time.Duration(delay))
		delayData, err := m.BatchGetDurationPointValue(ctx, pm, delayTime, isVirtual)
		if err != nil {
			log.Warnf("BatchGetDurationPointValue failed, pm: %v, t: %v, err: %v", pm, delayTime, err)
			continue
		} else if len(delayData) == 0 {
			log.Warnf("BatchGetDurationPointValue empty, pm: %v, t: %v", pm, t)
			continue
		}

		for p, item := range delayData {
			if _, ok := data[p]; !ok {
				data[p] = epoint.IntervalMap{}
			}
			for t, v := range item {
				data[p][t+delay] = v
			}
		}
	}

	return data, nil
}

// GetIntervalPointValue 获取间隔测点的数据
func (m *PointManager) GetIntervalPointValue(ctx context.Context, pm delayPoint, t time.Time) (epoint.ValueMap, error) {
	startTime := time.Now()
	defer func() {
		modcall.RecordDataQueryTime("DelaytimeRT", "interval", float64(time.Since(startTime).Milliseconds()))
	}()
	newDelay := pm.delay
	delayT := t.Add(-1 * time.Second * time.Duration(newDelay))
	// TODO 对接数据模块，完成测点数据查询
	// 使用pm.pointList和delay完成测点批量查询
	query := &repo.PointQueryReq{
		PointList: pm.pointList,
		Begin:     delayT.Unix(),
		End:       delayT.Unix(),
		Interval:  1,
	}
	pointValue, err := repo.GetPointDataSvc().GetPointDataTS(trpc.BackgroundContext(), query)
	if err != nil {
		err = fmt.Errorf("GetIntervalPointData failed, pm: %+v, time: %v; %w",
			pm, t, err)
		return nil, err
	}
	return pointValue, nil
}

// GetDurationPointValue TODO 对接数据模块，完成测点数据查询
// 使用pm.pointList和start, end完成测点批量查询
func (m *PointManager) GetDurationPointValue(ctx context.Context, pm delayPoint, t time.Time) (epoint.HistoryTimeValueMap, error) {
	startTime := time.Now()
	defer func() {
		modcall.RecordDataQueryTime("DelaytimeRT", "duration", float64(time.Since(startTime).Milliseconds()))
	}()
	newDelay := pm.delay
	start := t.Add(-1 * time.Second * time.Duration(newDelay))
	end := t
	// 按 interval 请求测点数据，反推批量获取的测点数量
	interval, err := GetDurationInterval(int64(end.Sub(start).Seconds()))
	if err != nil {
		return nil, err
	}
	query := &repo.PointQueryReq{
		PointList: pm.pointList,
		Begin:     start.Unix(),
		End:       end.Unix(),
		Interval:  int64(interval),
	}
	historyData, err := repo.GetPointDataSvc().GetPointDurationDataTS(
		trpc.BackgroundContext(), query)
	if err != nil {
		log.Errorf("GetDurationPointData failed, pm: %+v, start: %v, end: %v; %w",
			pm, start, end, err)
		return nil, err
	}
	allData := epoint.HistoryTimeValueMap{}
	for p, v := range historyData {
		allData[p] = v.InnerMap
	}
	return allData, nil
}
