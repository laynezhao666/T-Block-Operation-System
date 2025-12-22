package rmanager

import (
	"context"
	"sync"
	"time"

	"etrpc-go/log"

	"github.com/imdario/mergo"
	"github.com/panjf2000/ants/v2"

	"alarm-compute/conf"
	"alarm-compute/entity/epoint"
	"alarm-compute/logic/collector/validate"
	"alarm-compute/logic/point"
	"alarm-compute/logic/rules/rtask"
	"alarm-compute/repo"
	"alarm-compute/utils/common"
	"alarm-compute/utils/modcall"
)

type delayTask struct {
	pos  int
	task *rtask.RuleTask
}

// GetExecDealyTimeRuleTask 获取周期内需要执行的延迟策略
func (m *RuleManager) GetExecDealyTimeRuleTask(ctx context.Context, ts time.Time) map[string]*rtask.RuleTask {
	defer func() {
		endTime := time.Now()
		modcall.RecordDataChangeTime("DelaytimeRT", float64(endTime.Sub(ts).Milliseconds()))
	}()
	execRules := make(map[string]*rtask.RuleTask)
	totalPointList := m.DtPointRuleMap.GetTotalPointList()
	if len(totalPointList) == 0 {
		return execRules
	}
	batchSize := int(conf.ServerConf.DelayTimeConfig.VaryPointBatchSize)
	poolSize := int(conf.ServerConf.DelayTimeConfig.VaryPointPoolSize)
	if batchSize <= 0 {
		batchSize = 3000
		poolSize = 200
	}
	changedPointMap, err := repo.GetPointDataSvc().ParallelGetChangedPointMap(
		ctx, ts.Unix(), totalPointList, batchSize, poolSize)
	if err != nil {
		// 接口调用失败，打日志。同时运行全量策略
		log.Errorf("get changed point list failed, err: %v", err)
		m.DtTotalExecChan <- struct{}{}
		return execRules
	}
	for pointName, changedTs := range changedPointMap {
		if ruleKeyList, ok := m.DtPointRuleMap.Load(pointName); ok {
			for _, ruleKey := range ruleKeyList.([]string) {
				rule, ok := m.DelayTimeRule.Load(ruleKey)
				if !ok {
					continue
				}
				ruleTask := rule.(*rtask.RuleTask)
				timeLimit := conf.ServerConf.DelayTimeConfig.VaryPointQueryTimeSpan
				if _, ok := ruleTask.GetPointDelayMap()[pointName]; ok {
					timeLimit += ruleTask.GetPointDelayMap()[pointName]
				} else {
					log.Errorf("dt: failed to find point in pointDelayMap, point: %s", pointName)
				}
				if changedTs >= ts.Add(-time.Duration(timeLimit)*time.Second).Unix() {
					execRules[ruleKey] = ruleTask
				}
			}
		}
	}
	failedSet := validate.GetFailRuleCollector().GetFailedDt()
	for ruleKey := range failedSet {
		rule, ok := m.DelayTimeRule.Load(ruleKey)
		if ok {
			ruleTask := rule.(*rtask.RuleTask)
			execRules[ruleKey] = ruleTask
		}
	}
	return execRules
}

// StartDelayTimeRuleTask 延时策略计算
func (m *RuleManager) StartDelayTimeRuleTask(ctx context.Context) {
	log.Infof("compute delaytime rule task")
	var doTask = func(_ any) {
		m.doDelayTimeRuleTask(ctx, false)
	}
	var poolWg sync.WaitGroup
	execWpSzie := conf.ServerConf.DelayTimeConfig.ParallelExecWpSize
	wp, _ := ants.NewPoolWithFunc(int(execWpSzie), func(i interface{}) {
		doTask(i)
		poolWg.Done()
	}, ants.WithNonblocking(true))
	defer wp.Release()
	interval := conf.ServerConf.DelayTimeConfig.DelayTimeTaskInterval
	if interval == 0 {
		interval = 5
	}
	itv := time.Second * time.Duration(interval)
	tick := time.NewTicker(itv)
	var cycleCount int32
	for {
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
			if cycleCount == 0 {
				// 全量策略为同步执行，不使用协程池
				m.doDelayTimeRuleTask(ctx, true)
			} else {
				poolWg.Add(1)
				wpErr := wp.Invoke(cycleCount)
				if wpErr != nil {
					// 协程池满，为异常情况，为保障程序正常运行，不提交任务，打印错误日志
					poolWg.Done()
					log.AlarmContextf(ctx, "delaytime invoke failed, err: %v", wpErr)
				}
			}
			cycleCount++
			cycleCount %= int32(conf.ServerConf.DelayTimeConfig.TotalAnalyzeCycleCount)
		case <-m.DtTotalExecChan:
			cycleCount = 0
		}
	}
}

func (m *RuleManager) doDelayTimeRuleTask(ctx context.Context, isTotalAnalysis bool) {
	var rules map[string]*rtask.RuleTask
	now := time.Now()
	if isTotalAnalysis {
		rules = m.GetDelaytimeRules()
	} else {
		rules = m.GetExecDealyTimeRuleTask(ctx, now)
	}
	modcall.RecordAnalyzeTaskCnt("DelaytimeRT", isTotalAnalysis, len(rules))
	if !isTotalAnalysis {
		log.Infof("get changed delay rule, time %v", time.Since(now).Milliseconds())
	}
	m.EvalDelayTimeRuleTaskWithTime(ctx, now, isTotalAnalysis, rules)
}

// EvalDelayTimeRuleTaskWithTime 延时策略计算
func (m *RuleManager) EvalDelayTimeRuleTaskWithTime(ctx context.Context, now time.Time, isTotalAnalysis bool, rules map[string]*rtask.RuleTask) {
	log.Infof("start to eval delaytime rule task, len:%d, time %v, total analysis: %v", len(rules), now, isTotalAnalysis)
	if len(rules) == 0 {
		return
	}
	startAt := time.Now()
	defer func() {
		modcall.RecordStrategyTimeCost("DelaytimeRT", isTotalAnalysis, float64(time.Since(startAt).Milliseconds()))
	}()
	// 获取测点数据
	ts := now.Unix()
	pointValue, err := m.BatchGetPointValueDelayTime(ctx, rules, now, false)
	if err != nil {
		log.Errorf("BatchGetPointValueDelayTime failed, err: %v", err)
		pointValue = epoint.HistoryValueMap{}
	}
	failedTaskList := make([]string, len(rules))
	var doJobTask = func(i interface{}) error {
		dt := i.(*delayTask)
		rSuccess := dt.task.StartDelayTimeRuleTask(pointValue, ts)
		if !rSuccess {
			failedTaskList[dt.pos] = dt.task.GetKey()
		}
		return nil
	}
	// 执行任务
	poolSize := int(conf.ServerConf.DelayTimeConfig.DelayTimeTaskPoolSize)
	var poolWg sync.WaitGroup
	wp, _ := ants.NewPoolWithFunc(poolSize, func(i interface{}) {
		doJobTask(i)
		poolWg.Done()
	})
	defer wp.Release()
	var rIndex int
	for _, r := range rules {
		poolWg.Add(1)
		wp.Invoke(&delayTask{
			pos:  rIndex,
			task: r,
		})
		rIndex++
	}
	poolWg.Wait()
	// 打印失败的任务，每batch打印一次
	var failedTaskStr string
	batchSize := poolSize / 3
	var curBatch int
	for _, f := range failedTaskList {
		if len(f) > 0 {
			if len(failedTaskStr) > 0 {
				failedTaskStr += ","
			}
			failedTaskStr += f
			curBatch += 1
			if curBatch >= batchSize {
				curBatch = 0
				log.Warnf("dt failed due to the lack of points, list:%s", failedTaskStr)
				failedTaskStr = ""
			}
		}
	}
	if len(failedTaskStr) > 0 {
		log.Warnf("dt failed due to the lack of points, list:%s", failedTaskStr)
	}
}

// BatchGetPointValueDelayTime 延时测点 批量获取数据
func (m *RuleManager) BatchGetPointValueDelayTime(ctx context.Context, execRules map[string]*rtask.RuleTask, t time.Time, isValtual bool) (
	epoint.HistoryValueMap, error) {
	allData := epoint.HistoryValueMap{}
	mergeFn := func(newData epoint.HistoryValueMap) {
		for p, valueMap := range newData {
			if old, ok := allData[p]; ok {
				// 如果有相同测点，则合并数据
				mergo.Merge(&valueMap, old)
			}
			allData[p] = valueMap
		}
	}
	pm := m.DelayPointMap(execRules)
	pointDataList := make([]epoint.HistoryValueMap, 3)
	var localWg sync.WaitGroup
	getPointFn := func(index int, pointTypeFunc func(ctx context.Context, pm *epoint.DelayPointMap, t time.Time,
		isVartual bool) (epoint.HistoryValueMap, error)) {
		data, err := pointTypeFunc(ctx, pm, t, isValtual)
		if err == nil {
			pointDataList[index] = data
		} else {
			pointDataList[index] = epoint.HistoryValueMap{}
		}
	}
	localWg.Add(3)
	go func() {
		defer localWg.Done()
		getPointFn(0, m.getIntervalPoint)
	}()
	go func() {
		defer localWg.Done()
		getPointFn(1, m.getDurationPoint)
	}()
	go func() {
		defer localWg.Done()
		getPointFn(2, m.getRangePoint)
	}()
	localWg.Wait()
	for _, data := range pointDataList {
		mergeFn(data)
	}
	return allData, nil
}

// uniqueFn 获取全量数据后再进行去重
func uniqueFn(m map[int][]string) map[int][]string {
	uniqueMap := make(map[int][]string, len(m))
	for k, v := range m {
		uPointList := common.UniqueStringSlice(v)
		uniqueMap[k] = uPointList
	}

	return uniqueMap
}

// DelayPointMap 获取需要查询数据的测点
func (m *RuleManager) DelayPointMap(execRules map[string]*rtask.RuleTask) *epoint.DelayPointMap {
	pm := epoint.NewDelayPointMap()
	for _, rule := range execRules {
		rule.Alert.GetDelayPoints(pm)
		rule.Restore.GetDelayPoints(pm)
	}
	pm.HPointMap = uniqueFn(pm.HPointMap)
	pm.HDPointMap = uniqueFn(pm.HDPointMap)
	return pm
}

func (m *RuleManager) getIntervalPoint(ctx context.Context, pm *epoint.DelayPointMap, t time.Time,
	isVartual bool) (epoint.HistoryValueMap, error) {
	// 获取 interval 测点数据
	if len(pm.HPointMap) > 0 {
		data, err := point.GetPointManager().BatchGetIntervalPointValue(ctx, pm.HPointMap, t, isVartual)
		if err != nil {
			log.Warnf("BatchGetPointValue failed, hPointMap: %v, t: %v, err: %v", pm.HPointMap, t, err)
			return epoint.HistoryValueMap{}, err
		} else {
			if len(data) == 0 {
				log.Warnf("BatchGetPointValue empty, hPointMap: %v, t: %v", pm.HPointMap, t)
			}
			return data, nil
		}
	}
	return epoint.HistoryValueMap{}, nil
}

func (m *RuleManager) getDurationPoint(ctx context.Context, pm *epoint.DelayPointMap, t time.Time,
	isVirtual bool) (epoint.HistoryValueMap, error) {
	// 获取函数 duration 测点数据
	if len(pm.HDPointMap) > 0 {
		data, err := point.GetPointManager().BatchGetDurationPointValue(ctx, pm.HDPointMap, t, isVirtual)
		if err != nil {
			log.Warnf("BatchGetDurationPointValue failed, hdPointMap: %v, t: %v, err: %v", pm.HDPointMap, t, err)
			return epoint.HistoryValueMap{}, err
		} else {
			if len(data) == 0 {
				log.Warnf("BatchGetDurationPointValue empty, hdPointMap: %v, t: %v", pm.HDPointMap, t)
			}
			return data, nil
		}
	}
	return epoint.HistoryValueMap{}, nil
}

func (m *RuleManager) getRangePoint(ctx context.Context, pm *epoint.DelayPointMap, t time.Time,
	isVirtual bool) (epoint.HistoryValueMap, error) {
	if len(pm.HRPointMap) > 0 {
		data, err := point.GetPointManager().BatchGetRangePointValue(ctx, pm.HRPointMap, t, isVirtual)
		if err != nil {
			log.Warnf("BatchGetRangePointValue failed, HRPointMap: %v, t: %v, err: %v", pm.HRPointMap, t, err)
			return epoint.HistoryValueMap{}, err
		} else {
			if len(data) == 0 {
				log.Warnf("BatchGetRangePointValue empty, HRPointMap: %v, t: %v", pm.HRPointMap, t)
			}
			return data, nil
		}
	}
	return epoint.HistoryValueMap{}, nil
}
