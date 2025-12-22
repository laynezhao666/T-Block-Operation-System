// Package collector collector
package collector

import (
	"context"
	"sync"
	"time"

	"etrpc-go/log"

	"alarm-server/conf"
	"alarm-server/entity/model"
	"alarm-server/repo/store"
)

var (
	collector *ValidateColleror
	once      sync.Once
)

// RuleCollector 策略收集器
type RuleCollector struct {
	ValidWg sync.WaitGroup
	Mtx     sync.Mutex
	// 记录当前周期所有执行成功过的策略key及状态
	RecordRuleSet map[string]*model.ValidStoreData
}

// GetAndResetCollector GetAndResetCollector
func (rc *RuleCollector) GetAndResetCollector() map[string]*model.ValidStoreData {
	rc.ValidWg.Add(1)
	defer rc.ValidWg.Done()
	rc.Mtx.Lock()
	resRuleSet := rc.RecordRuleSet
	rc.RecordRuleSet = make(map[string]*model.ValidStoreData)
	rc.Mtx.Unlock()
	return resRuleSet
}

// BatchAddRecord 批量添加策略执行记录
func (rc *RuleCollector) BatchAddRecord(ts time.Time, recordMap map[string]*model.ValidStoreData) {
	rc.ValidWg.Wait()
	rc.Mtx.Lock()
	defer rc.Mtx.Unlock()
	if time.Since(ts) > time.Minute {
		// 处理时间太长，消息处理速度过慢
		log.Errorf("collector: collect time is too long, %v", time.Since(ts))
		return
	}
	for ruleKey, newValidItem := range recordMap {
		if storeItem, ok := rc.RecordRuleSet[ruleKey]; ok {
			if storeItem.EvalTime >= newValidItem.EvalTime {
				continue
			}
		}
		rc.RecordRuleSet[ruleKey] = newValidItem
	}
}

// ValidateColleror 收集器
type ValidateColleror struct {
	RealTimeCollector *RuleCollector
	DelayCollector    *RuleCollector
}

// GetValidateColleror 获取全局策略收集器
func GetValidateColleror() *ValidateColleror {
	once.Do(func() {
		collector = &ValidateColleror{
			RealTimeCollector: &RuleCollector{
				RecordRuleSet: make(map[string]*model.ValidStoreData),
			},
			DelayCollector: &RuleCollector{
				RecordRuleSet: make(map[string]*model.ValidStoreData),
			},
		}
	})
	return collector
}

// BatchAddRuleRecord 批量添加策略执行记录
func (v *ValidateColleror) BatchAddRuleRecord(recordMap map[string]*model.ValidStoreData,
	ridType int) {
	if ridType == 0 {
		go v.RealTimeCollector.BatchAddRecord(time.Now(), recordMap)
	} else if ridType == 1 {
		go v.DelayCollector.BatchAddRecord(time.Now(), recordMap)
	}
}

// RegularStoreValidMsg 定时同步至远程存储
func (v *ValidateColleror) RegularStoreValidMsg(ctx context.Context, wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()
	interval := conf.ServerConf.RuleValidConfig.RegularStoreInterval
	if interval <= 0 {
		interval = 5
	}
	itv := time.Second * time.Duration(interval)
	tick := time.NewTicker(itv)
	for {
		select {
		case <-ctx.Done():
			tick.Stop()
			return
		case <-tick.C:
			rtRecord := v.RealTimeCollector.GetAndResetCollector()
			dtRecord := v.DelayCollector.GetAndResetCollector()
			go store.GetRedisStoreApi().BatchStoreRuleRecord(rtRecord)
			store.GetRedisStoreApi().BatchStoreRuleRecord(dtRecord)
		}
	}
}
