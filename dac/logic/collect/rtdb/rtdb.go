// Package rtdb 提供门禁系统实时数据库的读写接口。
// 支持内存缓存和Redis两种存储后端。
package rtdb

import (
	"context"

	"dac/entity/model/rt"
	"dac/logic/dlm"
)

// GetPoints 获取点位数据。若持有分布式锁则从内存读取，否则从Redis读取。
func GetPoints(ctx context.Context, points rt.Points) error {
	if len(points) == 0 {
		return nil
	}

	if dlm.GetWorker().HasLock() {
		memoryInstance.GetPoints(points)
		return nil
	}

	return GetRedisPoints(ctx, points)
}

// SetPoints 设置点位数据到内存缓存
func SetPoints(points rt.Points, notify bool) {
	if len(points) == 0 {
		return
	}

	memoryInstance.SetPoints(points, notify)
}

// RegisterPointsUpdatedCallback 注册点位数据更新回调
func RegisterPointsUpdatedCallback(
	handler PointsUpdatedCallback, arg interface{},
) {
	memoryInstance.RegisterDataPointsUpdatedCallback(
		handler, arg)
}

// UnRegisterPointsUpdatedCallback 取消注册点位数据更新回调
func UnRegisterPointsUpdatedCallback(
	handler PointsUpdatedCallback,
) {
	memoryInstance.UnregisterDataPointsUpdatedCallback(
		handler)
}
