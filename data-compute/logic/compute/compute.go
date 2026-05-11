// Package compute 数据计算相关业务逻辑
package compute

import (
	"common/entity/consts"
	"common/util/expr"
	"context"
	"data-compute/entity/model"
	"data-compute/repo/report"
	"data-compute/repo/rpc"
	"data-compute/repo/store"
	"etrpc-go/log"
	"fmt"
	"strings"
	"sync"
	"time"
	pb "trpcprotocol/data-compute"

	"github.com/samber/lo"
	"trpc.group/trpc-go/trpc-go"
)

// IComputeApi 标准点计算策略相关实现接口
type IComputeApi interface {
	//ReceiveTask 接收调度器发送的策略
	ReceiveTask(req *pb.ReqReceiveTask)
	// ShowTask 查询计算任务
	ShowTask(ctx context.Context, req *pb.ReqQueryFilter) (*pb.RspShowTask, error)
	// ShowData 查询计算的数据
	ShowData(ctx context.Context, req *pb.ReqQueryFilter) (*pb.RspShowData, error)
	// StartCalcPoint 启动执行所有任务
	StartCalcPoint(ctx context.Context, wg *sync.WaitGroup)
}

var (
	computeApi IComputeApi
	initOnce   sync.Once
)

type localStrategy struct {
	stdPointName string                         // 测点名称
	mozuId       int32                          // 模组Id
	refVarMap    map[string]map[string]struct{} // 引用测点变量映射: 1307381448457609239.Tbat_1 -> A,B ，可能一个测点对应多个变量
	expression   string                         // 表达式
}

type computeApiImpl struct {
	strategyLock        sync.RWMutex                // 读写strategyMap锁
	strategyMap         map[string]*pb.PointTask    // 任务列表
	refPointStrategyMap map[string][]*localStrategy // 记录每个引用点->对应的标准点策略列表
	lastValueMap        sync.Map                    // 记录标准点上次计算的值
	cacheApi            rpc.ICacheQueryApi          // 数据查询接口
}

// GetComputeApi 创建默认的测点计算服务
func GetComputeApi() IComputeApi {
	initOnce.Do(func() {
		computeApi = &computeApiImpl{
			strategyMap: make(map[string]*pb.PointTask),
			cacheApi:    rpc.NewCacheQueryApi(),
		}
	})
	return computeApi
}

func getTaskKey(task *pb.PointTask) string {
	return fmt.Sprintf("%s.%s.%s", task.DeviceGid, task.PointNameEn, task.Version)
}

func (d *computeApiImpl) ReceiveTask(req *pb.ReqReceiveTask) {
	d.strategyLock.Lock()
	defer d.strategyLock.Unlock()
	if req.PublishType == 1 {
		clear(d.strategyMap)
		for _, task := range req.AddTask {
			d.strategyMap[getTaskKey(task)] = task
		}
	} else {
		for _, task := range req.DelTask {
			delete(d.strategyMap, getTaskKey(task))
		}
		for _, task := range req.AddTask {
			d.strategyMap[getTaskKey(task)] = task
		}
	}
	// 重新处理生成新的数据
	newTasks := lo.Values(d.strategyMap)
	newStrategyRefMap := make(map[string][]*localStrategy)
	for _, strategy := range newTasks {
		originRefs := strings.Split(strategy.ExpressionMap, ";")
		noBlankRefs := lo.Filter(originRefs, func(item string, index int) bool {
			return len(item) > 0
		})
		// 常量点
		if len(noBlankRefs) == 0 {
			existStrategyRefs, ok := newStrategyRefMap["constants"]
			if !ok {
				existStrategyRefs = make([]*localStrategy, 0)
			}
			newStrategyRefMap["constants"] = append(existStrategyRefs, &localStrategy{
				stdPointName: fmt.Sprintf("%s.%s", strategy.DeviceGid, strategy.PointNameEn),
				mozuId:       strategy.MozuId,
				refVarMap:    make(map[string]map[string]struct{}),
				expression:   strategy.Expression,
			})
			continue
		}
		// 转化为 测点->变量名称map
		newRefMap := make(map[string]map[string]struct{})
		for _, ref := range noBlankRefs {
			eqPos := strings.Index(ref, "=")
			varName, refPoint := ref[0:eqPos], ref[eqPos+1:]
			existRefs, ok := newRefMap[refPoint]
			if !ok {
				existRefs = make(map[string]struct{})
			}
			existRefs[varName] = struct{}{}
			newRefMap[refPoint] = existRefs
		}
		for refPoint := range newRefMap {
			existStrategyRefs, ok := newStrategyRefMap[refPoint]
			if !ok {
				existStrategyRefs = make([]*localStrategy, 0)
			}
			newStrategyRefMap[refPoint] = append(existStrategyRefs, &localStrategy{
				stdPointName: fmt.Sprintf("%s.%s", strategy.DeviceGid, strategy.PointNameEn),
				refVarMap:    newRefMap,
				mozuId:       strategy.MozuId,
				expression:   strategy.Expression,
			})
		}
	}
	d.refPointStrategyMap = newStrategyRefMap
}

func (d *computeApiImpl) ShowTask(ctx context.Context, req *pb.ReqQueryFilter) (*pb.RspShowTask, error) {
	d.strategyLock.RLock()
	defer d.strategyLock.RUnlock()
	strategies := lo.Values(d.strategyMap)
	if req.DeviceGid != "" {
		strategies = lo.Filter(strategies, func(item *pb.PointTask, index int) bool {
			return item.DeviceGid == req.DeviceGid
		})
	}
	if req.PointNameEn != "" {
		strategies = lo.Filter(strategies, func(item *pb.PointTask, index int) bool {
			return item.PointNameEn == req.PointNameEn
		})
	}
	if len(req.PointKey) > 0 {
		pointKeyMap := lo.SliceToMap(req.PointKey, func(item string) (string, struct{}) {
			return item, struct{}{}
		})
		strategies = lo.Filter(strategies, func(item *pb.PointTask, index int) bool {
			_, ok := pointKeyMap[fmt.Sprintf("%s.%s", item.DeviceGid, item.PointNameEn)]
			return ok
		})
	}
	res := lo.SliceToMap(strategies, func(item *pb.PointTask) (string, *pb.PointTask) {
		return fmt.Sprintf("%s.%s", item.DeviceGid, item.PointNameEn), item
	})
	return &pb.RspShowTask{
		Map:   res,
		Total: int32(len(res)),
	}, nil
}

func (d *computeApiImpl) ShowData(ctx context.Context, req *pb.ReqQueryFilter) (*pb.RspShowData, error) {
	res := make(map[string]*pb.PointValue)
	d.lastValueMap.Range(func(key, value any) bool {
		point := value.(*model.Point)
		pv := &pb.PointValue{
			Time:    point.Time,
			Value:   point.Value,
			Quality: point.Quality,
		}
		if len(req.PointKey) > 0 {
			if lo.Contains(req.PointKey, point.Name) {
				res[point.Name] = pv
			}
			return true
		}
		res[point.Name] = pv
		return true
	})
	return &pb.RspShowData{
		Map:   res,
		Total: int32(len(res)),
	}, nil
}

func (d *computeApiImpl) StartCalcPoint(ctx context.Context, wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()
	nowSec := time.Now().Unix()
	ticker := time.NewTicker(time.Second * 5)
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-ticker.C:
			// 60s整个周期计算一次
			if (msg.Unix()-nowSec)%60 == 0 {
				go d.calcFullPoints(wg)
			} else {
				go d.calcChangedPoints(msg.Add(-6*time.Second), wg)
			}
		}
	}
}

// calcChangedPoints 计算变化的测点
func (d *computeApiImpl) calcChangedPoints(begin time.Time, wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()
	// 查询短时间内变化的测点
	ctx := trpc.BackgroundContext()
	allRefPointNames := lo.Keys(d.refPointStrategyMap)
	needCalcStrategyMap := make(map[string]*localStrategy)
	changedPoints, err := d.cacheApi.ReadChanged(ctx, allRefPointNames, begin.Unix())
	if err != nil {
		// 调用接口失败，fallback为全量执行
		log.Errorf("read batch changed point err: %v", err)
		for _, relateStrategies := range d.refPointStrategyMap {
			for _, strategy := range relateStrategies {
				needCalcStrategyMap[strategy.stdPointName] = strategy
			}
		}
	} else {
		// 根据变化的测点，计算出需要计算的标准点
		for pointName := range changedPoints {
			if relateStrategies, ok := d.refPointStrategyMap[pointName]; ok {
				for _, strategy := range relateStrategies {
					needCalcStrategyMap[strategy.stdPointName] = strategy
				}
			}
		}
	}
	if len(needCalcStrategyMap) == 0 {
		return
	}
	_, calcChangedPoints := d.calcPoints(lo.Values(needCalcStrategyMap), consts.CalcTypeChanged)
	store.GetKafkaStore().BatchWrite(calcChangedPoints, consts.PointIntervalChanged)
}

func (d *computeApiImpl) calcFullPoints(wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()
	needCalcStrategyMap := make(map[string]*localStrategy)
	for _, relateStrategies := range d.refPointStrategyMap {
		for _, strategy := range relateStrategies {
			needCalcStrategyMap[strategy.stdPointName] = strategy
		}
	}
	fullPoints, _ := d.calcPoints(lo.Values(needCalcStrategyMap), consts.CalcTypePeriod)
	store.GetKafkaStore().BatchWrite(fullPoints, consts.PointIntervalPeriod)
}

func (d *computeApiImpl) calcPoints(needCalcPoints []*localStrategy, calcType string) ([]*model.Point, []*model.Point) {
	begin := time.Now()
	// 需要计算的标准点去重
	uniqueStdPoints := lo.UniqBy(needCalcPoints, func(item *localStrategy) string {
		return item.stdPointName
	})
	// 取出需要计算的标准点中所有涉及的引用测点
	uniqueRefPoints := make(map[string]struct{})
	for _, point := range uniqueStdPoints {
		for refPoint := range point.refVarMap {
			uniqueRefPoints[refPoint] = struct{}{}
		}
	}
	// 查询所有涉及的引用测点最新值
	refPointsVal, err := d.cacheApi.ReadLatest(trpc.BackgroundContext(), lo.Keys(uniqueRefPoints), begin.Unix())
	newPoints, changedPoints := d.evalPoints(uniqueStdPoints, refPointsVal, &begin)
	// 查询数据异常,所有测点质量设置为查询测点接口异常
	if err != nil {
		log.Errorf("read latest point data fail, err: %s", err.Error())
		for _, point := range newPoints {
			point.Quality = int32(consts.QualityQueryCacheApiErr)
		}
	}

	goodPoints := lo.Filter(newPoints, func(item *model.Point, index int) bool {
		return item.Quality == 0
	})

	// 上报相关指标
	costMs := time.Since(begin).Milliseconds()
	dimMetric := map[string]string{report.CalcTypeKey: calcType}
	report.PointCalcExpectCnt.ReportWithDim(float64(len(uniqueStdPoints)), dimMetric)
	report.PointCalcSuccessCnt.ReportWithDim(float64(len(goodPoints)), dimMetric)
	report.PointCalcCost.ReportWithDim(float64(costMs), dimMetric)

	return newPoints, changedPoints
}

func (d *computeApiImpl) evalPoints(uniqueStdPoints []*localStrategy,
	refPointsVal map[string]*model.Point, begin *time.Time) ([]*model.Point, []*model.Point) {
	// 保存所有新计算的测点
	newPoints := make([]*model.Point, 0, len(uniqueStdPoints))
	// 保存变化的测点
	changedPoints := make([]*model.Point, 0, len(uniqueStdPoints))
	for _, stdPoint := range uniqueStdPoints {
		variableValMap := make(map[string]any)
		var lessPoint []string
		// 查询策略依赖的测点是否都存在
		for pointName, variables := range stdPoint.refVarMap {
			if val, ok := refPointsVal[pointName]; ok && val.IsValid(begin.Unix()) {
				for variable := range variables {
					variableValMap[variable] = val.Value
				}
			} else {
				lessPoint = append(lessPoint, pointName)
			}
		}
		newPoint := &model.Point{
			Name:    stdPoint.stdPointName,
			Time:    begin.Unix(),
			EvalTms: begin.UnixMilli(),
			MozuId:  stdPoint.mozuId,
		}
		if len(lessPoint) > 0 {
			log.Warnf("point[%s] less point, ref point [%v] not found.", stdPoint.stdPointName, lessPoint)
			newPoint.Quality = int32(consts.QualityCalcLessPointErr)
		} else {
			val, qua, err := expr.EvalFloat(stdPoint.expression, variableValMap)
			if err != nil {
				log.Warnf("point[%s] eval fail, err: %s", stdPoint.stdPointName, err.Error())
			}
			newPoint.Quality = int32(qua)
			newPoint.Value = val
		}
		newPoints = append(newPoints, newPoint)
		// 保存并比对测点是否发生变化
		lastData, loaded := d.lastValueMap.LoadOrStore(stdPoint.stdPointName, newPoint)
		lastPoint := lastData.(*model.Point)
		if !loaded || (lastPoint.Value != newPoint.Value || lastPoint.Quality != newPoint.Quality) {
			*lastPoint = *newPoint
			changedPoints = append(changedPoints, newPoint)
		}
	}
	return newPoints, changedPoints
}
