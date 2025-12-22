package validate

import (
	"context"
	"sync"
	"time"

	"alarm-compute/conf"
)

// RuleCollector RuleCollector
type RuleCollector struct {
	FailedRtChan   chan string
	FailedDtChan   chan string
	RtFailedMap    map[string]struct{}
	DtFailedMap    map[string]struct{}
	RtDispatchChan chan map[string]struct{}
	DtDispatchChan chan map[string]struct{}
}

var (
	ruleCollector RuleCollector
	ruleOnce      sync.Once
)

// GetFailRuleCollector GetFailRuleCollector
func GetFailRuleCollector() *RuleCollector {
	ruleOnce.Do(func() {
		batchSize := conf.ServerConf.ValidateRecordConfig.BatchSize
		if batchSize < 1000 {
			batchSize = DefaultBatchSize
		}
		ruleCollector.FailedRtChan = make(chan string, batchSize)
		ruleCollector.FailedDtChan = make(chan string, batchSize)
		ruleCollector.RtDispatchChan = make(chan map[string]struct{}, 10)
		ruleCollector.DtDispatchChan = make(chan map[string]struct{}, 10)
		ruleCollector.RtFailedMap = make(map[string]struct{})
		ruleCollector.DtFailedMap = make(map[string]struct{})
	})
	return &ruleCollector
}

// AddFailedRt AddFailedRt
func (rc *RuleCollector) AddFailedRt(item string) {
	rc.FailedRtChan <- item
}

// AddFailedDt AddFailedDt
func (rc *RuleCollector) AddFailedDt(item string) {
	rc.FailedDtChan <- item
}

// GetFailedRt GetFailedRt
func (rc *RuleCollector) GetFailedRt() map[string]struct{} {
	select {
	case failedSet := <-rc.RtDispatchChan:
		return failedSet
	default:
		return map[string]struct{}{}
	}
}

// GetFailedDt GetFailedDt
func (rc *RuleCollector) GetFailedDt() map[string]struct{} {
	select {
	case failedSet := <-rc.DtDispatchChan:
		return failedSet
	default:
		return map[string]struct{}{}
	}
}

// CollectFailed CollectFailed
func (rc *RuleCollector) CollectFailed(ctx context.Context, wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()
	interval := conf.ServerConf.ValidateRecordConfig.FailedDispatchInterval
	if interval == 0 {
		interval = 5000
	}
	flushInterval := time.Duration(interval) * time.Millisecond
	ticker := time.NewTicker(flushInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case item := <-rc.FailedRtChan:
			rc.RtFailedMap[item] = struct{}{}
		case item := <-rc.FailedDtChan:
			rc.DtFailedMap[item] = struct{}{}
		case <-ticker.C:
			if len(rc.RtFailedMap) > 0 {
				rc.RtDispatchChan <- rc.RtFailedMap
				rc.RtFailedMap = make(map[string]struct{})
			}
			if len(rc.DtFailedMap) > 0 {
				rc.DtDispatchChan <- rc.DtFailedMap
				rc.DtFailedMap = make(map[string]struct{})
			}
		}
	}
}
