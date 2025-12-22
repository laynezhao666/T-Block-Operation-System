package worker

import (
	"agent/utils/osal"
	"time"
)

const (
	DefaultWorkerNum = 200
	MaxWorkerNum     = 500
	MaxCollectTime   = 10 * time.Minute
)

var (
	pool *WorkPool
)

// WorkPool 工作池
type WorkPool struct {
	sem *osal.Semaphore
}

func init() {
	pool = NewWorkPool(DefaultWorkerNum)
}

// NewWorkPool 新建工作池
func NewWorkPool(n int) *WorkPool {
	if n > MaxWorkerNum || n < 0 {
		n = MaxWorkerNum
	}
	p := &WorkPool{
		sem: osal.NewSemaphore(n),
	}
	for i := 0; i < n; i++ {
		p.sem.Post()
	}
	return p
}

// Acquire 获取一个 worker
func (w *WorkPool) Acquire() bool {
	if w == nil {
		return false
	}
	return w.sem.Wait(MaxCollectTime)
}

// Release 释放一个 worker
func (w *WorkPool) Release() {
	if w == nil {
		return
	}
	w.sem.Post()
}
