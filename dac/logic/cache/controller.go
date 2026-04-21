// Package cache 提供门禁控制器的内存缓存管理，包括控制器状态、告警和事件的增量拉取。
package cache

import (
	"context"
	"time"

	"dac/entity/config"
	"dac/entity/model/db"
	"dac/entity/model/rt"
	"dac/entity/utils"
	"dac/logic/cache/alarm"
	"dac/logic/cache/event"
	"dac/logic/cache/fetcher"
	"dac/logic/collect/dispatcher"
	"dac/repo/dac"
)

// 缓存刷新和fetcher名称常量
const (
	// refreshControllerTime 控制器缓存刷新间隔
	refreshControllerTime = time.Second * 20

	// alarmFetcherName 告警索引fetcher名称
	alarmFetcherName = "alarm index"
	// alarmTimestampFetcherName 告警时间戳fetcher名称
	alarmTimestampFetcherName = "alarm timestamp"
	// eventFetcherName 事件索引fetcher名称
	eventFetcherName = "event index"
	// eventTimestampFetcherName 事件时间戳fetcher名称
	eventTimestampFetcherName = "event timestamp"
)

// isDoorsEqual 判断两个门禁控制器下的门编号是否相同
func isDoorsEqual(lhs, rhs rt.DoorController) bool {
	lhsLen := len(lhs.Doors)
	rhsLen := len(rhs.Doors)
	if lhsLen != rhsLen {
		return false
	}

	lhsDoors := make(map[int]struct{}, lhsLen)
	for i := range lhs.Doors {
		lhsDoors[lhs.Doors[i].Number] = struct{}{}
	}
	for i := range rhs.Doors {
		if _, ok := lhsDoors[rhs.Doors[i].Number]; !ok {
			return false
		}
	}
	return true
}

// GetController 根据ID获取单个控制器缓存
func (c *Cache) GetController(id db.IDType) (rt.DoorController, bool) {
	c.controllerMutex.RLock()
	defer c.controllerMutex.RUnlock()

	controller, ok := c.controllers[id]
	return controller, ok
}

// GetControllers 根据ID列表批量获取控制器缓存
func (c *Cache) GetControllers(ids []db.IDType) map[db.IDType]rt.DoorController {
	c.controllerMutex.RLock()
	defer c.controllerMutex.RUnlock()
	controllers := make(map[db.IDType]rt.DoorController, len(ids))
	for _, id := range ids {
		controller, ok := c.controllers[id]
		if !ok {
			continue
		}
		controllers[id] = controller
	}

	return controllers
}

// GetAllControllers 获取所有控制器缓存的副本
func (c *Cache) GetAllControllers() map[db.IDType]rt.DoorController {
	c.controllerMutex.RLock()
	defer c.controllerMutex.RUnlock()

	controllers := make(map[db.IDType]rt.DoorController, len(c.controllers))
	for id, c := range c.controllers {
		controllers[id] = c
	}
	return controllers
}

// HasController 判断指定ID的控制器是否存在于缓存中
func (c *Cache) HasController(id db.IDType) bool {
	c.controllerMutex.RLock()
	defer c.controllerMutex.RUnlock()

	_, ok := c.controllers[id]
	return ok
}

// addAlarmFetchers 给门禁控制器添加告警fetcher
func (c *Cache) addAlarmFetchers(ctx context.Context, controllers []rt.DoorController) {
	var err error
	c.alarmFetcherMutex.Lock()
	defer c.alarmFetcherMutex.Unlock()

	for i := range controllers {
		if config.C.IgnoreFetch(controllers[i].MozuID) {
			config.Log.Infof("ignore fetch controller %+v alarm, mozu: %v", controllers[i], controllers[i].MozuID)
			continue
		}

		id := controllers[i].ID

		oldFetcher, hasOldFetcher := c.alarmFetchers[id]

		var newFetcher fetcher.Fetcher

		loopWaitTime := utils.GetAlarmFetchLoopWaitTime(controllers[i].Extend)
		waitTime := utils.GetAlarmFetchWaitTime(controllers[i].Extend)
		// excel导入的告警同步间隔，如果没有便设置默认的同步等待时间
		if utils.IsFetchByTimestamp(controllers[i].Extend) ||
			config.C.IsFetchByTime(controllers[i].MozuID, controllers[i].Channel.ID) {
			newFetcher = fetcher.NewByTimestamp(id, alarmTimestampFetcherName, loopWaitTime, waitTime,
				alarm.GetTimestamp, alarm.FetchByTimestamp, c.getControllerFun(id), nil, controllers[i].MozuID)
		} else {
			newFetcher = fetcher.NewByIndex(id, alarmFetcherName, loopWaitTime, waitTime,
				alarm.GetIndex, alarm.Fetch, c.getControllerFun(id), nil, controllers[i].MozuID)
		}

		if err = newFetcher.Start(ctx); err != nil {
			config.Log.Warnf("start fetch controller %v alarm error: %v", id, err)
			continue
		}

		if hasOldFetcher {
			oldFetcher.Stop()
		}

		c.alarmFetchers[id] = newFetcher
	}
}

// addEventFetchers 给门禁控制器添加事件fetcher
func (c *Cache) addEventFetchers(ctx context.Context, controllers []rt.DoorController) {
	var err error
	c.eventFetcherMutex.Lock()
	defer c.eventFetcherMutex.Unlock()

	for i := range controllers {
		if config.C.IgnoreFetch(controllers[i].MozuID) {
			config.Log.Infof("ignore fetch controller %+v event, mozu: %v", controllers[i], controllers[i].MozuID)
			continue
		}

		id := controllers[i].ID

		oldFetcher, hasOldFetcher := c.eventFetchers[id]

		var newFetcher fetcher.Fetcher

		loopWaitTime := utils.GetEventFetchLoopWaitTime(controllers[i].Extend)
		waitTime := utils.GetEventFetchWaitTime(controllers[i].Extend)

		// excel导入的事件同步间隔，如果没有便设置默认的同步等待时间
		if utils.IsFetchByTimestamp(controllers[i].Extend) ||
			config.C.IsFetchByTime(controllers[i].MozuID, controllers[i].Channel.ID) {
			newFetcher = fetcher.NewByTimestamp(id, eventTimestampFetcherName, loopWaitTime, waitTime,
				event.GetTimestamp, event.FetchByTimestamp, c.getControllerFun(id), func(ctx context.Context) interface{} {
					return c.GetCardStaffMap(c.getControllerFun(id)(ctx).MozuID)
				}, controllers[i].MozuID)
		} else {
			newFetcher = fetcher.NewByIndex(id, eventFetcherName, loopWaitTime, waitTime,
				event.GetIndex, event.Fetch, c.getControllerFun(id), func(ctx context.Context) interface{} {
					return c.GetCardStaffMap(c.getControllerFun(id)(ctx).MozuID)
				}, controllers[i].MozuID)
		}

		if err = newFetcher.Start(ctx); err != nil {
			config.Log.Warnf("start fetch controller %v events error: %v", id, err)
			continue
		}

		if hasOldFetcher {
			oldFetcher.Stop()
		}

		c.eventFetchers[id] = newFetcher
	}
}

// getControllerFun 返回获取指定ID控制器的闭包函数
func (c *Cache) getControllerFun(id db.IDType) fetcher.GetControllerFun {
	return func(ctx context.Context) rt.DoorController {
		return c.GetControllers([]db.IDType{id})[id]
	}
}

// deleteAlarmFetchers 停止并删除指定控制器的告警fetcher
func (c *Cache) deleteAlarmFetchers(ids []db.IDType) {
	c.alarmFetcherMutex.Lock()
	defer c.alarmFetcherMutex.Unlock()

	for _, id := range ids {
		f, ok := c.alarmFetchers[id]
		if !ok {
			continue
		}
		f.Stop()

		delete(c.alarmFetchers, id)
	}
}

// deleteEventFetchers 停止并删除指定控制器的事件fetcher
func (c *Cache) deleteEventFetchers(ids []db.IDType) {
	c.eventFetcherMutex.Lock()
	defer c.eventFetcherMutex.Unlock()

	for _, id := range ids {
		f, ok := c.eventFetchers[id]
		if !ok {
			continue
		}
		f.Stop()

		delete(c.eventFetchers, id)
	}
}

// addControllers 添加控制器到缓存并启动采集和fetcher
func (c *Cache) addControllers(ctx context.Context, controllers []rt.DoorController) {
	c.setControllers(controllers)
	go func() {
		dispatcher.Get().AddControllers(ctx, controllers)
	}()

	go c.addAlarmFetchers(ctx, controllers)
	go c.addEventFetchers(ctx, controllers)
}

// setControllers 将控制器列表写入缓存map
func (c *Cache) setControllers(controllers []rt.DoorController) {
	for i := range controllers {
		c.controllers[controllers[i].ID] = controllers[i]
	}
}

// deleteControllers 从缓存中删除控制器并停止相关采集
func (c *Cache) deleteControllers(ids []db.IDType) {
	for i := range ids {
		delete(c.controllers, ids[i])
	}

	go dispatcher.Get().DeleteControllers(ids)
	go c.deleteEventFetchers(ids)
	go c.deleteAlarmFetchers(ids)
}

// refreshController 检查并更新缓存中的门禁控制器
func (c *Cache) refreshController(ctx context.Context) {
	// 不需要筛选模组
	controllers, doors, err := dac.GetRW().GetAllDoorControllersAndDoors(ctx, "")
	if err != nil {
		config.Log.Warnf("get all door controllers and doors error: %v", err)
		return
	}

	controllerMap := make(map[db.IDType]db.DoorController, len(controllers))
	for i := range controllers {
		controllerMap[controllers[i].ID] = controllers[i]
	}

	toAddControllers := make([]rt.DoorController, 0, len(controllers))
	toDeleteControllers := make([]db.IDType, 0, len(controllers))
	toSetControllers := make([]rt.DoorController, 0, len(controllers))

	c.controllerMutex.Lock()

	for i := range c.controllers {
		id := c.controllers[i].ID
		_, ok := controllerMap[id]
		if !ok {
			// 若缓存中的门禁控制器已被删除
			// 则停止采集该控制器
			toDeleteControllers = append(toDeleteControllers, id)
		}
	}
	for i := range controllers {
		var newController rt.DoorController
		newController.DoorController = controllers[i]
		newController.Doors = doors[controllers[i].ID]

		id := newController.ID
		oldController, ok := c.controllers[id]
		if !ok || oldController.Version < newController.Version || !isDoorsEqual(oldController, newController) {
			// 若当前不存在该门禁控制器或门禁控制器有更新或门列表不一致
			// 则重新添加该控制器进行采集
			toAddControllers = append(toAddControllers, newController)
		} else {
			// 其他情况，需要更新门信息
			toSetControllers = append(toSetControllers, newController)
		}
	}

	c.setControllers(toAddControllers)
	c.setControllers(toSetControllers)
	c.deleteControllers(toDeleteControllers)
	c.controllerMutex.Unlock()

	c.addControllers(ctx, toAddControllers)
}

// refreshControllerLoop 定时刷新控制器缓存的主循环
func (c *Cache) refreshControllerLoop(ctx context.Context) {
	for {
		c.refreshController(ctx)
		select {
		case <-time.After(refreshControllerTime):
			break
		case <-ctx.Done():
			config.Log.Info("stop refresh controller loop.")
			return
		}
	}
}
