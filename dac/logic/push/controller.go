// Package push 实现门禁测点数据的周期推送功能。
package push

import (
	"context"
	"time"

	"dac/entity/config"
	"dac/entity/model/db"
	"dac/repo/dac"
)

// getController 根据ID从缓存中获取门禁控制器信息
func (w *worker) getController(id db.IDType) (db.DoorController, bool) {
	w.controllerMutex.RLock()
	defer w.controllerMutex.RUnlock()

	c, ok := w.controllers[id]
	return c, ok
}

// notifyGetController 异步通知获取指定ID的控制器信息
func (w *worker) notifyGetController(id db.IDType) {
	go func() {
		w.toGetControllerChan <- id
	}()
}

// getControllerLoop 监听控制器获取请求，按需刷新控制器缓存
func (w *worker) getControllerLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case id := <-w.toGetControllerChan:
			if _, ok := w.getController(id); ok {
				break
			}
			w.refreshController(ctx, id)
		}
	}
}

// refreshControllerLoop 定时刷新所有控制器缓存（每小时一次）
func (w *worker) refreshControllerLoop(ctx context.Context) {
	for {
		w.freshAllController(ctx)
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Hour):
			break
		}
	}
}

// freshAllController 从数据库获取所有门禁控制器并更新缓存
func (w *worker) freshAllController(ctx context.Context) {
	// 获取所有模组的门禁控制器
	cs, err := dac.GetRW().GetAllDoorControllers(ctx, "")
	if err != nil {
		config.Log.Warnf("get all controllers error: %v", err)
		return
	}

	w.controllerMutex.Lock()
	defer w.controllerMutex.Unlock()

	for i := range cs {
		w.controllers[cs[i].ID] = cs[i]
	}
}

// refreshController 从数据库获取单个控制器并更新缓存
func (w *worker) refreshController(ctx context.Context, id db.IDType) {
	c, err := dac.GetRW().GetDoorControllerRecord(ctx, id)
	if err != nil {
		config.Log.Warnf("get controller %v error: %v", id, err)
		return
	}

	w.controllerMutex.Lock()
	defer w.controllerMutex.Unlock()

	w.controllers[id] = c
}
