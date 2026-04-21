// Package cache 提供门禁控制器数据的本地缓存管理，
// 包括控制器信息缓存、时间同步和定时刷新等功能。
package cache

import (
	"context"
	"sync"
	"time"

	"dac/entity/config"
	"dac/entity/model/db"
	"dac/entity/model/driver"
	"dac/logic/collect/dispatcher"
	"dac/logic/dlm"
)

// syncTimeInterval 控制器时间同步间隔（每7天同步一次）
const (
	syncTimeInterval = time.Hour * 24 * 7
)

// syncTime 向所有控制器同步当前时间（需持有分布式锁）
func (c *Cache) syncTime(ctx context.Context) {
	if !dlm.GetWorker().HasLock() {
		config.Log.Infof("has no lock, do not sync time, sleeping...")
		return
	}

	controllers := c.GetAllControllers()

	var wg sync.WaitGroup

	// 并发向每个控制器发送时间同步请求
	for id := range controllers {
		wg.Add(1)
		go func(id db.IDType) {
			defer wg.Done()

			_, e := dispatcher.Get().DoSyncRequest(&db.Request{
				ControllerID: id,
				Method:       driver.MethodSetTime,
			})
			if e != nil {
				config.Log.Warnf("auto set controller %v time error: %v", id, e)
				return
			}
			config.Log.Infof("auto set controller %v time success", id)
		}(id)
	}

	wg.Wait()
}

// syncTimeLoop 定时循环执行控制器时间同步
func (c *Cache) syncTimeLoop(ctx context.Context) {
	for {
		c.syncTime(ctx)
		select {
		case <-time.After(syncTimeInterval):
			break
		case <-ctx.Done():
			return
		}
	}
}
