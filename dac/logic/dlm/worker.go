// Package dlm 提供分布式锁管理，确保同一时刻只有一个实例运行采集任务。
package dlm

import (
	"context"
	"sync/atomic"

	"dac/entity/consts"
	"dac/entity/redis"

	"github.com/google/uuid"
)

// w 全局Worker单例
var (
	w *Worker
)

// Worker 分布式锁工作器，通过Redis实现互斥
type Worker struct {
	id         string      // 唯一 id
	mutex      redis.Mutex // Redis分布式锁
	lockStatus int32       // 是否持有锁，0=未持有，1=持有
}

// Init 初始化分布式锁Worker并启动锁竞争循环
func Init(ctx context.Context) {
	w = newWorker()
	w.start(ctx)
}

// UnInit 释放分布式锁并停止Worker
func UnInit(ctx context.Context) {
	if w == nil {
		return
	}
	w.stop(ctx)
}

// newWorker 创建新的Worker实例
func newWorker() *Worker {
	w := new(Worker)

	atomic.StoreInt32(&w.lockStatus, 0)
	w.id = uuid.Must(uuid.NewUUID()).String()
	w.mutex = redis.GetMutex(
		consts.RedisLockName, w.id,
		consts.RedisLockExpireTime)

	w.Infof("redis mutex name: %v, value: %v",
		consts.RedisLockName, w.id)
	return w
}

// start 启动后台锁竞争协程
func (w *Worker) start(ctx context.Context) {
	go w.lockLoop(ctx)
}

// stop 释放锁并停止Worker
func (w *Worker) stop(ctx context.Context) {
	w.unlock(ctx)
}

// GetWorker 获取全局Worker实例
func GetWorker() *Worker {
	return w
}

// InitWithoutRedis 用于测试程序在没有 Redis 环境时初始化
// 创建一个默认持有锁的 Worker，跳过 Redis 依赖
func InitWithoutRedis() {
	w = &Worker{
		id:         uuid.Must(uuid.NewUUID()).String(),
		lockStatus: 1, // 默认持有锁
	}
}
