// Package dlm 提供基于Redis的分布式锁管理功能。
package dlm

import (
	"context"
	"sync/atomic"
	"time"

	"dac/entity/consts"
	"dac/entity/redis"
)

// HasLock 检查当前节点是否持有分布式锁
func (w *Worker) HasLock() bool {
	return atomic.LoadInt32(&w.lockStatus) > 0
}

// setHasLock 设置锁持有状态（原子操作）
func (w *Worker) setHasLock(b bool) {
	v := int32(0)
	if b {
		v = 1
	}
	atomic.StoreInt32(&w.lockStatus, v)
}

// lockLoop 定时获取锁或延长有效期
func (w *Worker) lockLoop(ctx context.Context) {
	for {
		w.Infof("lockLoop tick, hasLock: %v", w.HasLock())
		w.lockOrExtend(ctx)
		select {
		case <-ctx.Done():
			w.Infof("stop lockLoop.")
			return
		case <-time.After(consts.RedisLockExtendTime):
			break
		}
	}
}

// lockOrExtend 尝试获取锁或延长已持有锁的有效期
func (w *Worker) lockOrExtend(ctx context.Context) {
	// 若尚未获取到锁
	if !w.HasLock() {
		// 尝试获取锁
		err := w.mutex.LockContext(ctx)
		if err != nil {
			w.Warnf("try to get redis lock failed: %v", err)
			return
		}
		w.setHasLock(true)
		w.Infof("get redis lock: %v", w.id)
		if err := redis.GetClient().Set(ctx, consts.RedisKeyGetLockChannelIP, consts.ServiceIP, 0).Err(); err != nil {
			w.Warnf("set get lock channel ip in redis error, err: %s", err.Error())
		}
		return
	}

	// 尝试延长锁的有效期
	ok, err := w.mutex.ExtendContext(ctx)
	if err != nil || !ok {
		w.setHasLock(false)
		w.Warnf("extend lock failed: error = %v, ok = %v", err, ok)
		if err := redis.GetClient().Del(ctx, consts.RedisKeyGetLockChannelIP).Err(); err != nil {
			w.Warnf("del get lock channel ip in redis error, err: %s", err.Error())
		}
		return
	}
	w.Infof("extend lock success")
}

// unlock 释放分布式锁
func (w *Worker) unlock(ctx context.Context) {
	if !w.HasLock() {
		return
	}

	_, err := w.mutex.UnlockContext(ctx)
	if err != nil {
		w.Warnf("unlock redis error: %v", err)
	}
}
