// Package push 提供门禁测点数据的推送服务，支持周期推送和变化推送。
package push

import (
	"context"
	"sync"

	"dac/entity/model/db"
	"dac/entity/model/rt"
	"dac/logic/collect/rtdb"

	"github.com/robfig/cron/v3"
)

// www 全局推送Worker单例
var (
	www = newWorker()
)

// controllerPointsType 控制器ID到RTDB模型的映射
type controllerPointsType map[db.IDType]*rtdb.Model

// controllerChangedPointsType 控制器ID到变化点位的映射
type controllerChangedPointsType map[db.IDType]*changedPointsType

// worker 测点数据推送工作器
type worker struct {
	c *cron.Cron // 定时任务调度器

	// 控制器缓存
	controllers         map[db.IDType]db.DoorController
	controllerMutex     sync.RWMutex
	toGetControllerChan chan db.IDType

	// 周期推送的测点缓存
	controllerPointsCache      controllerPointsType
	controllerPointsCacheMutex sync.RWMutex

	// 变化推送的测点缓存
	firstAccessMutex             sync.RWMutex
	firstAccess                  map[string]struct{}
	controllerChangedPoints      controllerChangedPointsType
	controllerChangedPointsMutex sync.RWMutex
}

// GetWorker 获取全局推送Worker实例
func GetWorker() *worker {
	return www
}

// newWorker 创建新的推送Worker实例
func newWorker() *worker {
	w := new(worker)
	w.controllerPointsCache = make(controllerPointsType)
	w.c = cron.New(cron.WithSeconds())
	w.controllers = make(map[db.IDType]db.DoorController)
	w.toGetControllerChan = make(chan db.IDType, 100)

	w.firstAccess = make(map[string]struct{}, 200)
	w.controllerChangedPoints = make(
		controllerChangedPointsType, 200)

	return w
}

// Start 启动推送服务，包括周期推送和变化推送
func (w *worker) Start(ctx context.Context) error {
	_, err := w.c.AddFunc("0 * * * * *", func() {
		w.reportPeriod(ctx)
	})
	if err != nil {
		return err
	}

	w.c.Start()

	go w.refreshControllerLoop(ctx)
	go w.getControllerLoop(ctx)
	go w.reportChangedPointsLoop(ctx)

	return nil
}

// getControllerCache 获取控制器的RTDB缓存模型，不存在则自动创建
func (w *worker) getControllerCache(
	controllerID db.IDType,
) *rtdb.Model {
	w.controllerPointsCacheMutex.RLock()
	controllerCache, ok :=
		w.controllerPointsCache[controllerID]
	if ok {
		w.controllerPointsCacheMutex.RUnlock()
		return controllerCache
	}
	w.controllerPointsCacheMutex.RUnlock()

	w.controllerPointsCacheMutex.Lock()
	defer w.controllerPointsCacheMutex.Unlock()

	if controllerCache, ok =
		w.controllerPointsCache[controllerID]; !ok {
		controllerCache = rtdb.NewRtdbModel()
		controllerCache.RegisterDataPointsUpdatedCallback(
			w.changedCallback, controllerID)
		w.controllerPointsCache[controllerID] =
			controllerCache
	}
	return controllerCache
}

// SetPoints 设置控制器的测点数据到推送缓存
func (w *worker) SetPoints(
	controllerID db.IDType,
	points rt.Points, notify bool,
) {
	controllerCache := w.getControllerCache(controllerID)
	controllerCache.SetPoints(points, notify)
}
