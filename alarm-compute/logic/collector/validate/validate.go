// Package validate 策略生效收集器
package validate

import (
	"context"
	"sync"
	"time"

	pb "trpcprotocol/alarm-compute"

	"alarm-compute/conf"
	"alarm-compute/repo"
)

const (
	DefaultBatchSize = 1200
)

var (
	taskCollector ValidCollector
	once          sync.Once
)

// ValidCollector 任务生效收集器
type ValidCollector struct {
	//有效性任务 chan 实时上报
	RtValidator chan *pb.ValidateTaskItem
	DtValidator chan *pb.ValidateTaskItem
}

// GetTaskCollector GetTaskCollector
func GetTaskCollector() *ValidCollector {
	once.Do(func() {
		batchSize := conf.ServerConf.ValidateRecordConfig.BatchSize
		if batchSize < 1000 {
			batchSize = DefaultBatchSize
		}
		taskCollector.RtValidator = make(chan *pb.ValidateTaskItem, 10*batchSize)
		taskCollector.DtValidator = make(chan *pb.ValidateTaskItem, 10*batchSize)
	})
	return &taskCollector
}

// AddValidateRecord AddValidateRecord
func (tc *ValidCollector) AddValidateRecord(item *pb.ValidateTaskItem) {
	if item.RidType == 0 {
		tc.RtValidator <- item
	} else if item.RidType == 1 {
		tc.DtValidator <- item
	}
}

func (tc *ValidCollector) startValidate(ctx context.Context, validateCh chan *pb.ValidateTaskItem, interval, batchSize int) {
	flushInterval := time.Duration(interval) * time.Millisecond
	ticker := time.NewTicker(flushInterval)
	defer ticker.Stop()
	var results []*pb.ValidateTaskItem
	for {
		select {
		case <-ctx.Done():
			return
		case item := <-validateCh:
			results = append(results, item)
			if len(results) >= int(batchSize) {
				go repo.GetCkafka().SendRuleValidMsg(results)
				results = make([]*pb.ValidateTaskItem, 0)
			}
		case <-ticker.C:
			if len(results) > 0 {
				go repo.GetCkafka().SendRuleValidMsg(results)
				results = make([]*pb.ValidateTaskItem, 0)
			}
		}
	}
}

// StartValidate StartValidate
// 实时策略和延时策略分开消费，加快发送速率
func (tc *ValidCollector) StartValidate(ctx context.Context, wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()
	interval := conf.ServerConf.ValidateRecordConfig.FlushInterval
	batchSize := conf.ServerConf.ValidateRecordConfig.BatchSize
	if interval == 0 {
		interval = 1000
		batchSize = 1000
	}
	go tc.startValidate(ctx, tc.RtValidator, int(interval), int(batchSize))
	tc.startValidate(ctx, tc.DtValidator, int(interval), int(batchSize))
}
