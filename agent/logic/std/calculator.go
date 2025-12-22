package std

import (
	"agent/logic/distribution/base"
	"common/util/expr"
	"errors"
	"math/rand"
	"strconv"
	"sync"
	"time"

	"trpc.group/trpc-go/trpc-go/log"

	"agent/entity/config"
	"agent/entity/consts"
	"agent/entity/definition"
	model2 "agent/entity/model/data"
	"agent/logic/cm"
	"agent/logic/collector/device/model"
	"agent/logic/collector/rtdb"
	rtdbModel "agent/logic/collector/rtdb/model"
	"agent/logic/distribution/distributor"
	utils2 "agent/logic/distribution/distributor/utils"
	"agent/utils"
	"agent/utils/osal"
)

type calculator struct {
	collectID2StdInfo      map[definition.DataPointIDType][]model.StdInstancePointInfo //采集到标准映射的关系
	collectID2StdInfoMutex sync.RWMutex

	cpCache      map[definition.DataPointIDType]rtdbModel.DataPoint // 采集测点缓存
	cpCacheMutex sync.RWMutex

	stdPoints definition.DataPointIDsType //标准点列表

	stopped      bool
	wg           sync.WaitGroup
	distributors distributor.Distributors
}

// stdResult 用于封装计算结果
type stdResult struct {
	stdInfo model.StdInstancePointInfo
	val     interface{}
	quality consts.Quality
	tms     int64
}

var (
	instance *calculator
)

// GetCalManager 获取计算管理器
func GetCalManager() *calculator {
	return instance
}

// newCalManager 创建计算管理器
func newCalManager() *calculator {
	instance = &calculator{
		collectID2StdInfo: make(map[definition.DataPointIDType][]model.StdInstancePointInfo,
			consts.CollectPointChangeCount),
		cpCache:      make(map[definition.DataPointIDType]rtdbModel.DataPoint, consts.CollectPointChangeCount),
		distributors: base.GetDistributorList(consts.StdChange),
	}
	return instance
}

func callback(points rtdbModel.DataPoints, _ interface{}) interface{} {
	instance.cpCacheMutex.Lock()
	defer instance.cpCacheMutex.Unlock()

	for i := range points {
		p := &points[i]
		//if !p.IsValueChanged {
		//  未变化点不触发标准点计算
		//	continue
		//}
		instance.cpCache[p.ID] = *p
	}
	return nil
}

// Unload 卸载
func (cal *calculator) Unload() {
	cal.collectID2StdInfoMutex.Lock()
	defer cal.collectID2StdInfoMutex.Unlock()

	cal.cpCacheMutex.Lock()
	defer cal.cpCacheMutex.Unlock()

	cal.collectID2StdInfo = map[definition.DataPointIDType][]model.StdInstancePointInfo{}
	cal.cpCache = map[definition.DataPointIDType]rtdbModel.DataPoint{}

	// 从实时数据库中删除设备下的所有测点
	rtdb.ClearDataPoints(cal.stdPoints)
	cal.stdPoints = definition.DataPointIDsType{}
}

// Reload 重新加载
func (cal *calculator) Reload() error {
	cal.Unload()
	return cal.LoadConfig()
}

// LoadConfig 加载配置
func (cal *calculator) LoadConfig() error {
	cal.collectID2StdInfoMutex.Lock()
	defer cal.collectID2StdInfoMutex.Unlock()

	// 读取
	data := cm.Worker().GetStdData()
	if data == nil {
		return errors.New("std config empty")
	}

	copyData := data.Copy()
	// 解析生成 采集-》标准 的关系
	for _, v := range copyData.StdPointsInfo {
		mapping := mapping(v.Mapping)
		cp, err := mapping.getCollectPoints()
		if err != nil {
			continue
		}

		param, err := mapping.getMappingParam()
		if err != nil {
			continue
		}
		v.Param = param

		for _, p := range cp {
			cpId := definition.DataPointIDType(p)
			if std, ok := cal.collectID2StdInfo[cpId]; ok {
				std = append(std, v)
				cal.collectID2StdInfo[cpId] = std
			} else {
				cal.collectID2StdInfo[cpId] = []model.StdInstancePointInfo{v}
			}
		}

		pointName := v.StdDevice + consts.DefaultIDSep + v.StdPoint
		cal.stdPoints = append(cal.stdPoints, definition.DataPointIDType(pointName))
	}

	log.Infof("std point count:%v; related collect point count:%v", len(cal.stdPoints),
		len(cal.collectID2StdInfo))

	return nil
}

func (cal *calculator) start() {
	cal.LoadConfig()
	cal.stopped = false
	cal.wg.Add(1)
	go cal.loop()
}

func (cal *calculator) stop() {
	cal.stopped = true
	cal.wg.Wait()
}

func (cal *calculator) loop() {
	defer cal.wg.Done()

	for {
		select {
		case <-time.After(time.Millisecond * 100):
			if cal.stopped {
				return
			}
			cal.process()
		}
	}
}

// 采集测点对应的标准点列表
func (cal *calculator) cpRelatedStd(tempCpCache map[definition.DataPointIDType]rtdbModel.DataPoint) model.StdInstancePointsInfo {
	cal.collectID2StdInfoMutex.Lock()
	defer cal.collectID2StdInfoMutex.Unlock()

	stdInfos := model.StdInstancePointsInfo{}
	duplicateMap := map[string]int{}

	for cpId := range tempCpCache {
		// 是否需要标准化
		stdInfoArr, ok := cal.collectID2StdInfo[cpId]
		if !ok {
			continue
		}

		// 采集 =》 标准 可能1：N
		for _, stdInfo := range stdInfoArr {
			key := stdInfo.StdDevice + consts.DefaultIDSep + stdInfo.StdPoint
			if _, ok := duplicateMap[key]; !ok {
				stdInfos = append(stdInfos, stdInfo)
				duplicateMap[key] = 1
			}
		}
	}
	return stdInfos
}

func (cal *calculator) process() {
	// 1. 复制并清空缓存
	tempCpCache := cal.copyAndClearCache()

	// 2. 获取本批次要计算的标准点列表
	processStdList := cal.cpRelatedStd(tempCpCache)
	if len(processStdList) == 0 {
		return
	}

	// 3. 并发计算标准点
	results, durationTime := cal.processStdPoints(processStdList, tempCpCache)

	// 4. 收集结果，构造标准点数据
	stdPoints, successCount := cal.collectResults(results)

	// 5. 存储及上报
	cal.storeAndReport(stdPoints)

	// 6. 记录耗时及上报指标
	cal.reportMetrics(successCount, len(processStdList), durationTime)
}

// 复制并清空缓存，避免并发读写
func (cal *calculator) copyAndClearCache() map[definition.DataPointIDType]rtdbModel.DataPoint {
	cal.cpCacheMutex.Lock()
	defer cal.cpCacheMutex.Unlock()

	tempCpCache := make(map[definition.DataPointIDType]rtdbModel.DataPoint, len(cal.cpCache))
	for k, v := range cal.cpCache {
		tempCpCache[k] = v
	}
	cal.cpCache = make(map[definition.DataPointIDType]rtdbModel.DataPoint, consts.CollectPointChangeCount)
	return tempCpCache
}

// 并发计算标准点，返回结果切片
func (cal *calculator) processStdPoints(processStdList []model.StdInstancePointInfo,
	tempCpCache map[definition.DataPointIDType]rtdbModel.DataPoint) ([]stdResult, time.Duration) {
	startTime := time.Now()

	numWorkers := 3
	if config.GetRB().IsGatewayMode() {
		numWorkers = 20
	}

	taskCh := make(chan model.StdInstancePointInfo, len(processStdList))
	resultCh := make(chan stdResult, len(processStdList))
	var wg sync.WaitGroup

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for stdInfo := range taskCh {
				val, quality, tms := cal.stdEval(stdInfo, tempCpCache)
				resultCh <- stdResult{stdInfo, val, quality, tms}
			}
		}()
	}

	go func() {
		for _, stdInfo := range processStdList {
			taskCh <- stdInfo
		}
		close(taskCh)
	}()

	go func() {
		wg.Wait()
		close(resultCh)
	}()

	var results []stdResult
	for res := range resultCh {
		results = append(results, res)
	}

	durationTime := time.Since(startTime)
	return results, durationTime
}

// 收集结果，构造标准点数据，返回数据点切片和成功计数
func (cal *calculator) collectResults(results []stdResult) ([]rtdbModel.DataPoint, int) {
	stdId2Point := make(map[definition.DataPointIDType]rtdbModel.DataPoint, len(results))
	successCount := 0

	for _, res := range results {
		stdInfo := res.stdInfo
		val := res.val
		quality := res.quality
		tms := res.tms

		if config.GetRB().IsAlarmTestEnable() {
			if is, v := IsTestData(); is {
				val = v
				quality = consts.QualityOk
			}
		}

		if quality == consts.QualityOk {
			successCount++
		}

		if tms <= 0 {
			tms = utils.GetNowUTCTimeStamp()
		}

		rtData := rtdbModel.NewRTData()
		rtData.Val.Pv = osal.NewVariantWithValue(val)
		rtData.Val.Qua = quality
		rtData.Val.Tms = tms

		stdID := stdInfo.StdDevice + consts.DefaultIDSep + stdInfo.StdPoint
		stdPoint := rtdbModel.DataPoint{
			ID:             definition.DataPointIDType(stdID),
			DeviceGiD:      definition.DeviceGidType(definition.StdDevice),
			Rtd:            rtData,
			IsValueChanged: false,
			PointType:      definition.StdPointType,
		}
		stdId2Point[stdPoint.ID] = stdPoint
	}

	stdPoints := make([]rtdbModel.DataPoint, 0, len(stdId2Point))
	for _, v := range stdId2Point {
		stdPoints = append(stdPoints, v)
	}

	return stdPoints, successCount
}

// 存储数据点并异步上报
func (cal *calculator) storeAndReport(stdPoints []rtdbModel.DataPoint) {
	if len(stdPoints) == 0 {
		return
	}

	// 存入rtdb，自动赋值是否变化字段
	rtdb.SetDataPoints(stdPoints, false)

	// 变化上报
	changedPoints := getChangedPoints(stdPoints)
	if len(changedPoints) > 0 {
		go func(points rtdbModel.DataPoints) {
			cal.pushDataPoints(definition.StdDevice, points)
		}(changedPoints)
	}

	// 质量分析上报
	go func(points []rtdbModel.DataPoint) {
		cal.quaAnalysisReport(points)
	}(stdPoints)
}

// 上报指标
func (cal *calculator) reportMetrics(successCount, totalCount int, durationTime time.Duration) {
	cal.reportStdQuaMetrics(float64(successCount), float64(totalCount))
	cal.reportStdTimeMetrics(float64(durationTime))
}

func (cal *calculator) pushDataPoints(d definition.DeviceGidType, points rtdbModel.DataPoints) {
	data := &model2.DataUnit{
		DeviceGid: d,
		Points:    points,
	}
	arg := utils2.DistributorArgs{
		Time:     utils.GetNowUTCTime(),
		Interval: 1,
	}

	cal.distributors.BatchDistribute(data, &arg)
}

// IsTestData for 告警测试
func IsTestData() (bool, string) {
	nowUtc := utils.GetNowUTCTime()

	seconds := nowUtc.Second()
	if seconds > 30 {
		rand.Seed(time.Now().UnixNano())
		randomNumber := rand.Intn(10000)
		result := randomNumber + 100000
		resultString := strconv.Itoa(result)
		return true, resultString
	}

	return false, ""
}

func getChangedPoints(all rtdbModel.DataPoints) rtdbModel.DataPoints {
	var changedPoints rtdbModel.DataPoints
	for _, v := range all {
		if v.IsValueChanged {
			changedPoints = append(changedPoints, v)
		}
	}
	return changedPoints
}

func (cal *calculator) stdEval(stdInfo model.StdInstancePointInfo, tempCpCache map[definition.DataPointIDType]rtdbModel.DataPoint) (interface{}, consts.Quality, int64) {
	tms := int64(-1) // 采集时间
	// 变量映射
	parameters := make(map[string]interface{})

	for k, v := range stdInfo.Param {
		var rtdVal rtdbModel.RTValue
		collectPointID := definition.DataPointIDType(v)

		// 从临时缓存获取采集点数据
		if p, exists := tempCpCache[collectPointID]; exists {
			rtdVal = p.Rtd.Val
		} else {
			// 缓存未命中时从实时数据库获取
			dataPoints := rtdb.GetDataPointsByID([]definition.DataPointIDType{collectPointID})
			if len(dataPoints) != 1 {
				return nil, consts.QualityStdParamErr, tms
			}
			rtdVal = dataPoints[0].Rtd.Val
		}

		if !rtdVal.IsOK() {
			return nil, rtdVal.Qua, tms
		}

		fv, err := rtdVal.Pv.AsFloat()
		if err != nil {
			return nil, consts.QualityValueTypeError, tms
		}
		parameters[k] = float64(fv)
		if rtdVal.Tms > tms {
			tms = rtdVal.Tms
		}
	}

	result, quality, err := expr.EvalStr(stdInfo.Expr, parameters)
	if err != nil {
		log.Debugf("std calculation error | expr:%s | params:%+v | err:%v", stdInfo.Expr, parameters, err)
	}
	return result, consts.Quality(quality), tms
}
