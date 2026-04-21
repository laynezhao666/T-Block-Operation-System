// Package push 提供门禁测点数据的推送服务，支持周期推送和变化推送。
package push

import (
	"context"
	"dac/entity/model/db"
	"dac/entity/model/rt"
	"dac/entity/utils/ttime"
	"dac/logic/dlm"
	pb "dac/repo/pb/tcommon_point_data"
	"trpc.group/trpc-go/trpc-go/log"
)

// getAllPoints 获取所有控制器的测点数据快照
func (w *worker) getAllPoints() map[db.IDType]rt.Points {
	w.controllerPointsCacheMutex.RLock()
	defer w.controllerPointsCacheMutex.RUnlock()

	points := make(map[db.IDType]rt.Points, len(w.controllerPointsCache))
	for controllerID, rtdb := range w.controllerPointsCache {
		points[controllerID] = rtdb.GetAllPoints()
	}

	return points
}

// reportPeriod 执行一次周期性测点数据上报
func (w *worker) reportPeriod(ctx context.Context) {
	if !dlm.GetWorker().HasLock() {
		log.Warnf("没有获得锁，无法上报测点数据")
		return
	}

	t := ttime.GetNowUTC().UnixMilli()
	points := w.getAllPoints()

	w.reportAllControllerPoints(ctx, t, pb.DataKind_Period, points)
}

// CleanControllerPointsCache 清空指定控制器的测点缓存
func (w *worker) CleanControllerPointsCache(ids []db.IDType) {
	w.controllerPointsCacheMutex.Lock()
	defer w.controllerPointsCacheMutex.Unlock()

	for _, id := range ids {
		if rtdb, ok := w.controllerPointsCache[id]; ok {
			rtdb.DeleteAllPoints()
			delete(w.controllerPointsCache, id)
		}
	}
}
