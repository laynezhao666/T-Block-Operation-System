package rmanager

import (
	"context"
	"sync"
	"time"

	"etrpc-go/log"

	"github.com/panjf2000/ants/v2"

	"alarm-compute/conf"
	"alarm-compute/logic/collector/validate"
	"alarm-compute/logic/point"
	"alarm-compute/logic/rules/rtask"
	"alarm-compute/repo"
	"alarm-compute/utils/common"
	"alarm-compute/utils/modcall"
)

// GetExecRealTimeRuleTask 获取周期内需要执行的实时策略
func (m *RuleManager) GetExecRealTimeRuleTask(ctx context.Context, ts time.Time) map[string]*rtask.RuleTask {
	defer func() {
		endTime := time.Now()
		modcall.RecordDataChangeTime("RealtimeRT", float64(endTime.Sub(ts).Milliseconds()))
	}()
	execRules := make(map[string]*rtask.RuleTask)
	totalPointList := m.RtPointRuleMap.GetTotalPointList()
	if len(totalPointList) == 0 {
		return execRules
	}
	timeSpan := conf.ServerConf.RealTimeConfig.VaryPointQueryTimeSpan
	batchSize := int(conf.ServerConf.RealTimeConfig.VaryPointBatchSize)
	poolSize := int(conf.ServerConf.RealTimeConfig.VaryPointPoolSize)
	if timeSpan <= 0 {
		timeSpan = 3
		batchSize = 3000
		poolSize = 200
	}
	begin, end := ts.Add(-time.Duration(timeSpan)*time.Second).Unix(), ts.Unix()
	changedPointList, err := repo.GetPointDataSvc().
		ParallelGetChangedPointList(ctx, totalPointList, batchSize, poolSize, begin, end)
	if err != nil {
		log.Errorf("get changed point list failed, err: %v", err)
		m.RtTotalExecChan <- struct{}{}
		return execRules
	}
	for _, pointName := range changedPointList {
		if ruleKeyList, ok := m.RtPointRuleMap.Load(pointName); ok {
			for _, ruleKey := range ruleKeyList.([]string) {
				rule, ok := m.RealTimeRule.Load(ruleKey)
				if ok {
					ruleTask := rule.(*rtask.RuleTask)
					execRules[ruleKey] = ruleTask
				}
			}
		}
	}
	failedSet := validate.GetFailRuleCollector().GetFailedRt()
	for ruleKey := range failedSet {
		rule, ok := m.RealTimeRule.Load(ruleKey)
		if ok {
			ruleTask := rule.(*rtask.RuleTask)
			execRules[ruleKey] = ruleTask
		}
	}
	return execRules
}

// StartRealTimeRuleTask 实时策略计算
func (m *RuleManager) StartRealTimeRuleTask(ctx context.Context) {
	log.Infof("compute realtime rule task")
	var doTask = func(_ any) {
		m.doRealTimeRuleTask(ctx, false)
	}
	var poolWg sync.WaitGroup
	execWpSzie := conf.ServerConf.RealTimeConfig.ParallelExecWpSize
	wp, _ := ants.NewPoolWithFunc(int(execWpSzie), func(i interface{}) {
		doTask(i)
		poolWg.Done()
	}, ants.WithNonblocking(true))
	defer wp.Release()
	interval := conf.ServerConf.RealTimeConfig.RealTimeTaskInterval
	if interval == 0 {
		interval = 1
	}
	itv := time.Second * time.Duration(interval)
	tick := time.NewTicker(itv)
	var cycleCount int32
	for {
		select {
		case <-ctx.Done():
			poolWg.Wait()
			return
		case <-tick.C:
			if cycleCount == 0 {
				// 全量策略为同步执行，不使用协程池
				m.doRealTimeRuleTask(ctx, true)
			} else {
				poolWg.Add(1)
				err := wp.Invoke(cycleCount)
				if err != nil {
					// 协程池满，为异常情况，为保障程序运行，继续执行。打印错误日志
					poolWg.Done()
					log.AlarmContextf(ctx, "realtime invoke failed, err: %v", err)
				}
			}
			cycleCount++
			cycleCount %= int32(conf.ServerConf.RealTimeConfig.TotalAnalyzeCycleCount)
		case <-m.RtTotalExecChan:
			cycleCount = 0
		}
	}
}

func (m *RuleManager) doRealTimeRuleTask(ctx context.Context, isTotalAnalysis bool) {
	var rules map[string]*rtask.RuleTask
	now := time.Now()
	if isTotalAnalysis {
		rules = m.GetRealtimeRules()
	} else {
		rules = m.GetExecRealTimeRuleTask(ctx, now)
	}
	modcall.RecordAnalyzeTaskCnt("RealtimeRT", isTotalAnalysis, len(rules))
	if !isTotalAnalysis {
		log.Infof("get changed real rule, time %v", time.Since(now).Milliseconds())
	}
	m.EvalRealTimeRuleTaskWithTime(ctx, now, isTotalAnalysis, rules)
}

// ChunkRuleTask ChunkRuleTask
func (m *RuleManager) ChunkRuleTask(rules map[string]*rtask.RuleTask, size int) []map[string]*rtask.RuleTask {
	i := 0
	chunk := make([]map[string]*rtask.RuleTask, 0)
	tmp := make(map[string]*rtask.RuleTask, 0)
	cnt := len(rules)
	for key, rule := range rules {
		i++
		tmp[key] = rule
		if i%size == 0 {
			chunk = append(chunk, tmp)
			tmp = make(map[string]*rtask.RuleTask, 0)
		} else if i == cnt {
			chunk = append(chunk, tmp)
		}
	}
	return chunk
}

// EvalRealTimeRuleTaskWithTime 实时策略计算
func (m *RuleManager) EvalRealTimeRuleTaskWithTime(ctx context.Context, now time.Time, isTotalAnalysis bool, rules map[string]*rtask.RuleTask) {
	log.Infof("start to eval realtime rule task, len:%d, time %v, total analysis:%v", len(rules), now, isTotalAnalysis)
	if len(rules) == 0 {
		return
	}
	startAt := time.Now()
	defer func() {
		modcall.RecordStrategyTimeCost("RealtimeRT", isTotalAnalysis, float64(time.Since(startAt).Milliseconds()))
	}()
	chunkSize := int(conf.ServerConf.RealTimeConfig.RealTimeBatchSize)
	if chunkSize == 0 {
		chunkSize = 100
	}
	chunk := m.ChunkRuleTask(rules, chunkSize)
	var doJobTask = func(i interface{}) error {
		chunkRules := i.(map[string]*rtask.RuleTask)
		pointValue, err := m.getPointValueRealTime(ctx, chunkRules, now)
		if err != nil {
			log.Errorf("BatchGetPointValueRealTime failed, err: %v", err)
			pointValue = map[string]float64{}
		}
		failedTaskList := make([]string, len(chunkRules))
		var wg sync.WaitGroup
		var rIndex int
		for _, rule := range chunkRules {
			wg.Add(1)
			go func(index int, r *rtask.RuleTask, pv map[string]float64, wgp *sync.WaitGroup) {
				defer wgp.Done()
				rSuccess := r.StartRealtimeByData(pv, now.Unix())
				if !rSuccess {
					failedTaskList[index] = r.GetKey()
				}
			}(rIndex, rule, pointValue, &wg)
			rIndex++
		}
		wg.Wait()
		var failedStr string
		for _, f := range failedTaskList {
			if len(f) > 0 {
				if len(failedStr) > 0 {
					failedStr += ","
				}
				failedStr += f
			}
		}
		if len(failedStr) > 0 {
			log.Warnf("rt failed due to the lack of points, list:%s", failedStr)
		}
		return nil
	}
	var poolWg sync.WaitGroup
	wp, _ := ants.NewPoolWithFunc(int(conf.ServerConf.RealTimeConfig.RealTimeTaskPoolSize), func(i interface{}) {
		doJobTask(i)
		poolWg.Done()
	})
	defer wp.Release()
	for _, crules := range chunk {
		poolWg.Add(1)
		wp.Invoke(crules)
	}
	poolWg.Wait()
}

// BatchGetPointValueRealTime 实时测点，批量获取数据
func (m *RuleManager) getPointValueRealTime(ctx context.Context, execRules map[string]*rtask.RuleTask, t time.Time) (map[string]float64, error) {
	// TODO
	// 使用execRules所有使用到的测点，查询测点实时数据
	startTime := time.Now()
	defer func() {
		modcall.RecordDataQueryTime("RealtimeRT", "interval", float64(time.Since(startTime).Milliseconds()))
	}()
	pointList := []string{}
	for _, rule := range execRules {
		pointList = append(pointList, rule.GetPointList()...)
	}
	pointList = common.UniqueStringSlice(pointList)
	pointValue, err := point.GetPointManager().BatchGetRTPointValue(ctx, pointList, t)
	if err != nil {
		return nil, err
	}
	return pointValue, nil
}
