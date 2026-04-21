// Package delta 提供测点数据变化检测和增量写入Redis的功能。
package delta

import (
	"context"
	"sync"
	"time"

	"dac/entity/config"
	"dac/entity/model/rt"
	"dac/logic/collect/rtdb"
)

// maxNodeNum 增量数据链表最大节点数
const (
	maxNodeNum = 100
)

// w 全局增量数据Worker单例
var (
	w = &worker{
		first: make(map[string]struct{}),
		data:  New(maxNodeNum),
	}
)

// callback 测点数据更新回调，将变化的测点加入增量队列
func callback(points rt.Points, _ interface{}) interface{} {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	for i := range points {
		p := &points[i]
		// 值未变化且已首次上报过，则跳过
		if !p.IsValueChanged {
			if _, has := w.first[p.ID]; has {
				continue
			}
			w.first[p.ID] = struct{}{}
		}
		w.data.PushPoint(p)
	}

	return nil
}

// worker 增量数据处理工作器
type worker struct {
	mutex sync.RWMutex
	first map[string]struct{} // 首次上报标记
	data  *List               // 增量数据链表
}

// Start 启动增量数据上报和缓存刷新协程
func (w *worker) Start(ctx context.Context) {
	go w.reportLoop(ctx)
	go w.refreshLoop(ctx)
}

// refresh 清空首次上报标记，使所有测点重新上报一次
func (w *worker) refresh() {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	w.first = make(map[string]struct{})
}

// refreshLoop 定时清空首次上报标记
func (w *worker) refreshLoop(ctx context.Context) {
	for {
		w.refresh()
		select {
		case <-time.After(time.Hour):
			break
		case <-ctx.Done():
			config.Log.Infof("stop refresh delta cache")
			return
		}
	}
}

// processPoints 将测点数据批量写入Redis
func (w *worker) processPoints(
	ctx context.Context, points rt.Points,
) {
	newCtx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()

	rtdb.SetRedisPoints(newCtx, points)
}

// process 处理增量队列中的所有测点数据
func (w *worker) process(ctx context.Context) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	l := w.data.Len()
	if l == 0 {
		return
	}

	// 合并所有待写入的测点，一次性写入 Redis，避免并发启动大量 goroutine
	allPoints := make(rt.Points, 0)
	var tempNode *Element
	for e := w.data.Front(); e != nil; {
		points := e.Value
		tempNode = e.Next()
		w.data.Remove(e)
		e = tempNode
		allPoints = append(allPoints, points...)
	}

	if len(allPoints) > 0 {
		go w.processPoints(ctx, allPoints)
	}
}

// reportLoop 后台循环处理增量数据
func (w *worker) reportLoop(ctx context.Context) {
	for {
		w.process(ctx)
		select {
		case <-time.After(time.Second):
			break
		case <-ctx.Done():
			return
		}
	}
}
