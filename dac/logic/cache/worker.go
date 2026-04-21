// Package cache 提供门禁系统的运行时数据缓存，包括控制器、事件和告警等。
package cache

import (
	"context"
	"sync"
	"time"

	"dac/entity/model/db"
	"dac/entity/model/rt"
	"dac/logic/cache/fetcher"
)

// refreshRequestTime 请求刷新间隔
const (
	refreshRequestTime = time.Second * 10
)

// cache 全局缓存单例
var (
	cache = Cache{
		controllers:   make(map[db.IDType]rt.DoorController),
		eventFetchers: make(map[db.IDType]fetcher.Fetcher),
		alarmFetchers: make(map[db.IDType]fetcher.Fetcher),
	}
)

// Get 获取全局缓存实例
func Get() *Cache {
	return &cache
}

// Cache 运行时数据缓存，管理控制器、事件和告警拉取器
type Cache struct {
	controllerMutex sync.RWMutex
	controllers     map[db.IDType]rt.DoorController // 控制器缓存

	eventFetcherMutex sync.RWMutex
	eventFetchers     map[db.IDType]fetcher.Fetcher // 事件拉取器

	alarmFetcherMutex sync.RWMutex
	alarmFetchers     map[db.IDType]fetcher.Fetcher // 告警拉取器

	mozuCardStaffMutex sync.RWMutex
	mozuCardStaffMap   map[string]map[string]db.Staff // 模组卡号到员工的映射
}

// Start 启动缓存的后台刷新协程
func (c *Cache) Start(ctx context.Context) {
	go c.refreshControllerLoop(ctx)
	go c.refreshRequestLoop(ctx)
	go c.refreshCardStaffMapLoop(ctx)
	go c.syncTimeLoop(ctx)
}
