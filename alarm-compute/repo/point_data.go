package repo

import (
	"context"
	"sync"

	"etrpc-go/log"
	"etrpc-go/util/traceutil"
	dataPb "trpcprotocol/data-query"

	"github.com/panjf2000/ants/v2"
	"trpc.group/trpc-go/trpc-go"

	"alarm-compute/utils/common"
)

const (
	DefaultMaxDelaySecond = 2400
)

var (
	once         sync.Once
	pointDataSvc *PointDataSvc
)

// PointDataSvc 测点数据查询服务
type PointDataSvc struct {
	cli dataPb.DataClientProxy
}

// PointQueryReq 测点查询请求体
type PointQueryReq struct {
	PointList []string
	Begin     int64
	End       int64
	Interval  int64
}

type pointChangeQueryCon struct {
	// 要查询的测点列表
	pointList []string
	// 查询开始时间
	begin int64
	// 查询结束时间
	end int64
	// 当前索引位置，用于写回
	index int
}

// GetPointDataSvc GetPointDataSvc
func GetPointDataSvc() *PointDataSvc {
	once.Do(func() {
		pointDataSvc = &PointDataSvc{
			cli: dataPb.NewDataClientProxy(),
		}
	})
	return pointDataSvc
}

func (getter *PointDataSvc) getPointData(ctx context.Context, query *PointQueryReq) (*dataPb.QueryResponse, error) {
	req := &dataPb.QueryRequest{
		PointList: query.PointList,
		Begin:     query.Begin,
		End:       query.End,
		Interval:  query.Interval,
	}
	rsp, err := getter.cli.DataQuery(ctx, req)
	if err != nil {
		return nil, err
	}
	return rsp, nil
}

// GetPointDataTS 根据unix时间戳（ts）读取测点数据
func (getter *PointDataSvc) GetPointDataTS(ctx context.Context, query *PointQueryReq) (
	map[string]float64, error) {
	reqCtx := common.TracesCustomSpanDemo(ctx)
	rsp, err := getter.getPointData(reqCtx, query)
	if err != nil {
		log.Errorf("GetPointDataTS rpc err %v", err)
		return nil, err
	}
	pointValueMap := make(map[string]float64)
	missPoint := rsp.MissPointName
	for pointName, pointData := range rsp.GetPointData {
		if d, ok := pointData.InnerMap[query.Begin]; !ok {
			missPoint = append(missPoint, pointName)
		} else {
			pointValueMap[pointName] = d
		}
	}
	if len(missPoint) > 0 {
		log.Warnf("GetPointDataTS missPoint %v of time %d, traceId:%s", missPoint, query.Begin, traceutil.GetTraceID(reqCtx))
	}
	return pointValueMap, nil
}

// GetPointDurationDataTS 根据unix时间戳（ts）读取测点数据
func (getter *PointDataSvc) GetPointDurationDataTS(ctx context.Context, query *PointQueryReq) (
	map[string]*dataPb.InnerMap, error) {
	reqCtx := common.TracesCustomSpanDemo(ctx)
	rsp, err := getter.getPointData(reqCtx, query)
	if err != nil {
		log.Errorf("GetPointDurationDataTS rpc err %v", err)
		return nil, err
	}
	missPoint := rsp.MissPointName
	if len(missPoint) > 0 {
		log.Warnf("GetPointDurationDataTS missPoint %v of req:%v, traceId:%s", missPoint, query, traceutil.GetTraceID(reqCtx))
	}
	return rsp.GetPointData, nil
}

// GetChangedPointMap 获取测点最近一次的变化时间
func (getter *PointDataSvc) GetChangedPointMap(ctx context.Context, ts int64, pointList []string) (map[string]int64, error) {
	req := &dataPb.ChangeRequest{
		PointList: pointList,
		Begin:     ts - DefaultMaxDelaySecond,
	}
	rsp, err := getter.cli.DataChange(trpc.BackgroundContext(), req)
	if err != nil {
		return nil, err
	}
	return rsp.ChangedPointMap, nil
}

// ParallelGetChangedPointMap 并发获取变更测点
func (getter *PointDataSvc) ParallelGetChangedPointMap(ctx context.Context,
	ts int64, pointList []string, batchSize, poolSize int) (map[string]int64, error) {
	chunkList, err := common.ChunkStringList(pointList, batchSize)
	if err != nil {
		return nil, err
	}
	var retList = make([]map[string]int64, len(chunkList))
	var doQueryFunc = func(pos int, item []string) {
		chunkPointValue, err := getter.GetChangedPointMap(trpc.BackgroundContext(), ts, item)
		if err != nil {
			retList[pos] = make(map[string]int64)
		} else {
			retList[pos] = chunkPointValue
		}
	}
	var poolWg sync.WaitGroup
	wp, _ := ants.NewPoolWithFunc(poolSize, func(i interface{}) {
		query := i.(pointChangeQueryCon)
		doQueryFunc(int(query.index), query.pointList)
		poolWg.Done()
	})
	defer wp.Release()
	for index, chunk := range chunkList {
		poolWg.Add(1)
		queryCon := pointChangeQueryCon{
			pointList: chunk,
			index:     index,
		}
		wp.Invoke(queryCon)
	}
	poolWg.Wait()
	ret := make(map[string]int64)
	for _, v := range retList {
		for k, v := range v {
			ret[k] = v
		}
	}
	return ret, nil
}

// GetChangedPointList 获取一段时间内测点变化列表
func (getter *PointDataSvc) GetChangedPointList(ctx context.Context, pointList []string, begin, end int64) ([]string, error) {
	req := &dataPb.PointChangeRequest{
		PointList: pointList,
		Begin:     begin,
		End:       end,
	}
	rsp, err := getter.cli.DataPointChange(trpc.BackgroundContext(), req)
	if err != nil {
		return nil, err
	}
	return rsp.PointList, nil
}

// ParallelGetChangedPointList 并发获取变更测点
func (getter *PointDataSvc) ParallelGetChangedPointList(ctx context.Context,
	pointList []string, batchSize int, poolSize int, begin, end int64) ([]string, error) {
	chunkList, err := common.ChunkStringList(pointList, batchSize)
	if err != nil {
		return nil, err
	}
	var retList = make([][]string, len(chunkList))
	// 定义查询变化测点列表函数，用于协程池并发查询
	var doQueryFunc = func(pos int, item []string) {
		chunkPointValue, err := getter.GetChangedPointList(trpc.BackgroundContext(), item, begin, end)
		if err != nil {
			retList[pos] = make([]string, 0)
		} else {
			retList[pos] = chunkPointValue
		}
	}
	var poolWg sync.WaitGroup
	// 新建协程池
	wp, _ := ants.NewPoolWithFunc(poolSize, func(i interface{}) {
		query := i.(pointChangeQueryCon)
		doQueryFunc(int(query.index), query.pointList)
		poolWg.Done()
	})
	defer wp.Release()
	for index, chunk := range chunkList {
		poolWg.Add(1)
		wp.Invoke(pointChangeQueryCon{
			index:     index,
			pointList: chunk,
		})
	}
	poolWg.Wait()
	ret := retList[0]
	for _, v := range retList[1:] {
		ret = append(ret, v...)
	}
	return ret, nil
}
