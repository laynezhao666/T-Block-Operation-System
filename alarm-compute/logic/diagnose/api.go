package diagnose

import (
	"context"
	"fmt"
	"sync"
	"time"

	"alarm-compute/entity"
	"alarm-compute/entity/epoint"
	"alarm-compute/logic/point"
	"alarm-compute/logic/pointeval"
	"alarm-compute/logic/rules/rtask"
	"alarm-compute/repo"
	"etrpc-go/config"
	"etrpc-go/log"
	pb "trpcprotocol/alarm-compute"
	dataPb "trpcprotocol/data-query"

	"github.com/panjf2000/ants/v2"
)

// 告警诊断 / 回放
var (
	diagnoseSvc *diagnoseSvcImpl
	once        sync.Once
)

// IDiagnoseSvc IDiagnoseSvc
type IDiagnoseSvc interface {
	ExpCompute(ctx context.Context, req *pb.ReqExpCompute) (rsp *pb.RspExpCompute, err error)
}

// NewDiagnoseSvc NewDiagnoseSvc
func NewDiagnoseSvc() IDiagnoseSvc {
	once.Do(func() {
		diagnoseSvc = &diagnoseSvcImpl{
			taskChan: make(chan struct{}, config.GetInt32OrDefault("diagnose.channel_size", 200)),
		}
	})
	return diagnoseSvc
}

type diagnoseSvcImpl struct {
	taskChan chan struct{}
}

func (d *diagnoseSvcImpl) addTask() error {
	select {
	case d.taskChan <- struct{}{}:
		return nil
	default:
		return fmt.Errorf("诊断任务队列已满，请稍后再试")
	}
}

func (d *diagnoseSvcImpl) delTask() {
	<-d.taskChan
}

// ExpCompute
/*
1、 校验是否通道已满，不再处理
2、 加载测点数据
3、 执行诊断
*/
func (d *diagnoseSvcImpl) ExpCompute(ctx context.Context, req *pb.ReqExpCompute) (*pb.RspExpCompute, error) {
	if err := d.addTask(); err != nil {
		return nil, err
	}
	defer d.delTask()
	ptList := []*pointeval.PointTypeMap{}
	// 加载表达式
	pm := epoint.NewDelayPointMap()
	for _, item := range req.List {
		alarmTask := rtask.NewAlarmTask(rtask.AlarmTaskType)
		alarmTask.SetExp(item.Express, &entity.VariableGidMap{
			ExprMap: item.PMap,
		})
		err := alarmTask.UpdatePointFetchList()
		if err != nil {
			return nil, fmt.Errorf("表达式解析错误：%s, Err: %s", item.Express, err.Error())
		}
		ptList = append(ptList, alarmTask.Exp)
		if !req.WithValue {
			// 未附带测点值，需要计算测点延迟时间，用于获取测点数据
			alarmTask.GetDelayPoints(pm)
		}
	}
	// load测点数据
	var pv epoint.HistoryTimeValueMap
	var dataErr error
	if !req.WithValue {
		pv, dataErr = d.getPvForDiagnose(ctx, req, pm)
		if dataErr != nil {
			return nil, fmt.Errorf("获取测点数据失败,请求体：%v, Err: %s", req, dataErr.Error())
		}
	}
	// 计算结果
	res := &pb.RspExpCompute{
		List: make([]*pb.RspExpCompute_Item, len(req.List)),
	}
	var doExpTask = func(i interface{}) error {
		index := i.(int)
		itemPt := ptList[index]
		itemRes := d.evalExpWithPV(ctx, req, index, itemPt, pv)
		res.List[index] = itemRes
		return nil
	}
	poolSize := config.GetInt32OrDefault("diagnose.exp_pool_size", 5)
	var poolWg sync.WaitGroup
	wp, _ := ants.NewPoolWithFunc(int(poolSize), func(i interface{}) {
		doExpTask(i)
		poolWg.Done()
	})
	defer wp.Release()
	for i := range ptList {
		poolWg.Add(1)
		index := i
		wp.Invoke(index)
	}
	poolWg.Wait()
	return res, nil
}

func (d *diagnoseSvcImpl) evalExpWithPV(ctx context.Context, req *pb.ReqExpCompute, pos int,
	pt *pointeval.PointTypeMap, pv epoint.HistoryTimeValueMap) *pb.RspExpCompute_Item {
	res := &pb.RspExpCompute_Item{
		Express: pt.Express,
		CalRes:  make([]*pb.RspExpCompute_Item_StepResult, ((req.EndTime-req.BeginTime)/int64(req.Interval))+1),
	}
	var getPointDataTsMap = func(point_id string) map[int64]float64 {
		if req.WithValue {
			res, ok := req.List[pos].Pv[point_id]
			if !ok {
				return nil
			}
			return res.Tv
		} else {
			res, ok := pv[point_id]
			if !ok {
				return nil
			}
			return res
		}
	}
	var doStepTask = func(i interface{}) error {
		index := i.(int)
		evalTime := req.BeginTime + int64(index)*int64(req.Interval)
		itemRes := &pb.RspExpCompute_Item_StepResult{
			Timestamp: time.Unix(evalTime, 0).Format(time.DateTime),
			Success:   false,
			Fired:     false,
		}
		res.CalRes[index] = itemRes
		// 将epoint.HistoryTimeValueMap 根据evalTime转化为 将epoint.HistoryValueMap
		historyValueMap := epoint.HistoryValueMap{}
		for point_id := range pt.PointFetchList {
			pointTsMap := getPointDataTsMap(point_id)
			if pointTsMap == nil {
				// 测点数据缺失，本次不计算，直接返回失败结果
				return nil
			}
			historyValueMap[point_id] = epoint.IntervalMap{}
			for ts, v := range pointTsMap {
				if ts > evalTime {
					continue
				}
				historyValueMap[point_id][int(evalTime-ts)] = v
			}
		}
		result, err := pt.EvalWithIntervalPointData(historyValueMap)
		if err != nil {
			log.Errorf("告警诊断：表达式计算失败,表达式：%s, 计算时间：%d, Err: %s", pt.Express, evalTime, err.Error())
		} else {
			itemRes.Success = true
			itemRes.Fired = result.(bool)
		}
		return nil
	}
	// 并发计算结果
	poolSize := config.GetInt32OrDefault("diagnose.step_pool_size", 10)
	var poolWg sync.WaitGroup
	wp, _ := ants.NewPoolWithFunc(int(poolSize), func(i interface{}) {
		doStepTask(i)
		poolWg.Done()
	})
	defer wp.Release()
	for i := 0; req.BeginTime+int64(i)*int64(req.Interval) <= req.EndTime; i++ {
		poolWg.Add(1)
		index := i
		wp.Invoke(index)
	}
	poolWg.Wait()
	return res
}

// getPointDataForDiagnose 加载测点数据，用于告警诊断
// out: map[string]map[int64]float  map[point_id]map[time_unix]value
func (d *diagnoseSvcImpl) getPvForDiagnose(
	ctx context.Context, req *pb.ReqExpCompute, pm *epoint.DelayPointMap) (epoint.HistoryTimeValueMap, error) {
	// 协程池定义
	poolSize := config.GetInt32OrDefault("diagnose.data_pool_size", 10)
	dataChan := make(chan map[string]*dataPb.InnerMap, 2*poolSize)
	var getDataFunc = func(i interface{}) error {
		query := i.(*repo.PointQueryReq)
		qData, err := repo.GetPointDataSvc().GetPointDurationDataTS(ctx, query)
		if err != nil {
			return err
		} else {
			dataChan <- qData
		}
		return nil
	}
	var poolWg sync.WaitGroup
	wp, _ := ants.NewPoolWithFunc(int(poolSize), func(i interface{}) {
		getDataFunc(i)
		poolWg.Done()
	})
	defer wp.Release()
	// 处理并发获取的测点数据
	var dataHandleWg sync.WaitGroup
	dataHandleWg.Add(1)
	ret := epoint.HistoryTimeValueMap{}
	go func() {
		defer dataHandleWg.Done()
		for item := range dataChan {
			for pointId, tsMap := range item {
				if _, ok := ret[pointId]; !ok {
					ret[pointId] = map[int64]float64{}
				}
				for ts, v := range tsMap.InnerMap {
					ret[pointId][ts] = v
				}
			}
		}
	}()
	// 并发获取测点数据
	// 查询历史单个时间点测点数据，将回放用到的多个时间段合并查询
	for interval, points := range pm.HPointMap {
		query := d.geneDataQueryReq(req, interval, 0, points)
		poolWg.Add(1)
		wp.Invoke(query)
	}
	// 查询历史某个时间段的测点数据，将回放用到的多个时间段合并查询
	for duration, points := range pm.HDPointMap {
		query := d.geneDataQueryReq(req, 0, duration, points)
		poolWg.Add(1)
		wp.Invoke(query)
	}
	// 查询跳变函数相关测点数据
	for delay, durationMap := range pm.HRPointMap {
		for duration, points := range durationMap {
			query := d.geneDataQueryReq(req, delay, duration, points)
			poolWg.Add(1)
			wp.Invoke(query)
		}
	}
	go func() {
		poolWg.Wait()
		close(dataChan)
	}()
	dataHandleWg.Wait()
	return ret, nil
}

func (d *diagnoseSvcImpl) geneDataQueryReq(req *pb.ReqExpCompute, delay, duration int, points []string) *repo.PointQueryReq {
	var query *repo.PointQueryReq
	if duration == 0 {
		// 查询单个历史时间段的测点数据
		query = &repo.PointQueryReq{
			PointList: points,
			Begin:     req.BeginTime - int64(delay),
			End:       req.EndTime - int64(delay),
			Interval:  int64(req.Interval),
		}
	} else {
		// 查询一段历史时间的测点数据
		interval, err := point.GetDurationInterval(int64(duration))
		if err != nil {
			return nil
		}
		query = &repo.PointQueryReq{
			PointList: points,
			Begin:     req.BeginTime - int64(delay) - int64(duration),
			End:       req.EndTime - int64(delay),
			Interval:  int64(interval),
		}
	}
	return query
}
