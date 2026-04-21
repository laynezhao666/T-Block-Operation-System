// Package request 提供异步请求的定时清理功能。
package request

import (
	"context"
	"time"

	"dac/entity/config"
	"dac/entity/consts"
	"dac/repo/dac"
)

// cleanTime 清理任务执行间隔
const (
	cleanTime = time.Hour * 24
)

// w 全局Worker单例
var (
	w = new(Worker)
)

// Worker 异步请求清理工作器
type Worker struct {
}

// getWorker 获取全局Worker实例
func getWorker() *Worker {
	return w
}

// Init 初始化请求清理模块并启动后台清理协程
func Init(ctx context.Context) {
	getWorker().start(ctx)
}

// start 启动后台清理协程
func (w *Worker) start(ctx context.Context) {
	go w.cleanRequestsLoop(ctx)
}

// cleanRequests 执行一次请求清理，删除过期记录并标记超时记录
func (w *Worker) cleanRequests(ctx context.Context) {
	// 1. 计算时间阈值
	now := time.Now().UnixMilli()
	expirationThreshold := now - int64(time.Duration(config.C.ExpirationTime)*time.Hour*24/time.Millisecond)
	deletionThreshold := now - int64(time.Duration(config.C.DeletionTime)*time.Hour*24/time.Millisecond)

	config.Log.Infof("开始清理请求，过期阈值：%v天，删除阈值：%v天", config.C.ExpirationTime, config.C.DeletionTime)

	// 2. 先执行删除操作（删除更久远的记录）
	deletedCount, err := dac.GetRW().DeleteRequestsByTime(ctx, deletionThreshold)
	if err != nil {
		config.Log.Warnf("删除过期请求失败: %v", err)
		return
	}

	// 3. 再执行状态更新操作（标记过期但未到删除时间的记录）
	expiredCount, err := dac.GetRW().OutdatedRequestsByTime(ctx, expirationThreshold, consts.StateToBeExecuted)
	if err != nil {
		config.Log.Warnf("更新过期请求状态失败: %v", err)
		return
	}

	config.Log.Infof("清理请求完成 -- 删除记录数: %d, 过期记录数: %d", deletedCount, expiredCount)
}

// cleanRequestsLoop 后台循环执行请求清理任务
func (w *Worker) cleanRequestsLoop(ctx context.Context) {
	for {
		w.cleanRequests(ctx)
		select {
		case <-time.After(cleanTime):
			break
		case <-ctx.Done():
			config.Log.Infof("stop clean expired requests loop.")
			return
		}
	}
}
