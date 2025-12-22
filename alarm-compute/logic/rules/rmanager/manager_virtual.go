package rmanager

import (
	"context"
	"sync"
	"time"

	"etrpc-go/log"

	"github.com/panjf2000/ants/v2"

	"alarm-compute/conf"
	"alarm-compute/logic/rules/rtask"
	"alarm-compute/repo"
	"alarm-compute/utils/modcall"
)

// GetExecVirtualRuleTask 获取需要执行的虚拟点策略计算任务
func (m *RuleManager) GetExecVirtualRuleTask(ctx context.Context, ts time.Time) map[string]*rtask.RuleTask {
	defer func() {
		endTime := time.Now()
		modcall.RecordDataChangeTime("VirtualRT", float64(endTime.Sub(ts).Milliseconds()))
	}()
	execRules := make(map[string]*rtask.RuleTask)
	totalPointList := m.VtPointRuleMap.GetTotalPointList()
	if len(totalPointList) == 0 {
		return execRules
	}
	batchSize := int(conf.ServerConf.VirtualConfig.VaryPointBatchSize)
	poolSize := int(conf.ServerConf.VirtualConfig.VaryPointPoolSize)
	if batchSize <= 0 {
		batchSize = 3000
		poolSize = 200
	}
	changedPointMap, err := repo.GetPointDataSvc().ParallelGetChangedPointMap(
		ctx, ts.Unix(), totalPointList, batchSize, poolSize)
	if err != nil {
		// 接口调用失败，打日志。同时运行全量策略
		log.Errorf("get changed point list failed, err: %v", err)
		m.VtTotalExecChan <- struct{}{}
		return m.GetVirtualRules()
	}
	for pointName, changedTs := range changedPointMap {
		if ruleKeyList, ok := m.VtPointRuleMap.Load(pointName); ok {
			for _, ruleKey := range ruleKeyList.([]string) {
				rule, ok := m.VirtualRule.Load(ruleKey)
				if !ok {
					continue
				}
				ruleTask := rule.(*rtask.RuleTask)
				timeLimit := conf.ServerConf.VirtualConfig.VaryPointQueryTimeSpan
				if _, ok := ruleTask.GetPointDelayMap()[pointName]; ok {
					timeLimit += ruleTask.GetPointDelayMap()[pointName]
				} else {
					log.Errorf("vt: failed to find point in pointDelayMap, point: %s", pointName)
				}
				if changedTs >= ts.Add(-time.Duration(timeLimit)*time.Second).Unix() {
					execRules[ruleKey] = ruleTask
				}
			}
		}
	}
	if len(execRules) == 0 {
		m.EmptyVaryVtPointCount++
		if m.EmptyVaryVtPointCount >= int(conf.ServerConf.VirtualConfig.EmptyVaryPointCountLimit) {
			m.EmptyVaryVtPointCount = 0
			// 提前执行全量策略，刷新全量策略执行周期
			log.Warnf("empty changed virtual point leads to total analysis")
			m.VtTotalExecChan <- struct{}{}
			return execRules
		}
		return execRules
	}
	return execRules
}

// StartVirtualRuleTask 虚拟点策略计算
func (m *RuleManager) StartVirtualRuleTask(ctx context.Context) {
	log.Infof("compute virtualPoint rule task")
	interval := conf.ServerConf.VirtualConfig.VirtualTaskInterval
	if interval == 0 {
		interval = 2
	}
	itv := time.Second * time.Duration(interval)
	tick := time.NewTicker(itv)
	cycleCount := 0
	isTaskRunning := false
	for {
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
			if !isTaskRunning {
				isTaskRunning = true
				go func() {
					defer func() { isTaskRunning = false }()
					if cycleCount == 0 {
						m.doVirtualRuleTask(ctx, true)
					} else {
						m.doVirtualRuleTask(ctx, false)
					}
					cycleCount++
					cycleCount %= int(conf.ServerConf.VirtualConfig.TotalAnalyzeCycleCount)
				}()
				modcall.RecordAnalyzeDelayCnt("VirtualRT", false)
			} else {
				modcall.RecordAnalyzeDelayCnt("VirtualRT", true)
			}
		case <-m.VtTotalExecChan:
			cycleCount = 0
		}
	}
}

func (m *RuleManager) doVirtualRuleTask(ctx context.Context, isTotalAnalysis bool) {
	var rules map[string]*rtask.RuleTask
	now := time.Now()
	if isTotalAnalysis {
		rules = m.GetVirtualRules()
	} else {
		rules = m.GetExecVirtualRuleTask(ctx, now)
	}
	modcall.RecordAnalyzeTaskCnt("VirtualRT", isTotalAnalysis, len(rules))
	startAt := time.Now()
	defer func() {
		modcall.RecordStrategyTimeCost("VirtualRT", isTotalAnalysis, float64(time.Since(startAt).Milliseconds()))
	}()
	m.EvalVirtualRuleTaskWithTime(ctx, now, rules)
}

// EvalVirtualRuleTaskWithTime 虚拟点策略计算
func (m *RuleManager) EvalVirtualRuleTaskWithTime(ctx context.Context, now time.Time, rules map[string]*rtask.RuleTask) {
	log.Infof("start to eval virtual rule task, len:%d, time %v", len(rules), now)
	if len(rules) == 0 {
		return
	}
	pointValue, err := m.BatchGetPointValueDelayTime(ctx, rules, now, true)
	if err != nil {
		log.Errorf("VT BatchGetPointValue failed, err: %v", err)
		return
	}
	var doJobTask = func(i interface{}) error {
		r := i.(*rtask.RuleTask)
		r.StartVirtualRuleTask(pointValue, now)
		return nil
	}
	poolSize := conf.ServerConf.VirtualConfig.VirtualTaskPoolSize
	var poolWg sync.WaitGroup
	wp, _ := ants.NewPoolWithFunc(int(poolSize), func(i interface{}) {
		doJobTask(i)
		poolWg.Done()
	})
	defer wp.Release()
	for _, r := range rules {
		poolWg.Add(1)
		wp.Invoke(r)
	}
	poolWg.Wait()
}
