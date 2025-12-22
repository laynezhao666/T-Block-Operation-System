// Package lcache 缓存
package lcache

import (
	"fmt"
	"sync"

	"github.com/avast/retry-go"
	"trpc.group/trpc-go/trpc-database/localcache"
)

var (
	lcache *LocalCache
	once   sync.Once
)

const (
	// 活动告警缓存模版
	activeTmp = "active:%s"
)

// LocalCache 本地缓存
type LocalCache struct {
	localcache.Cache
}

// GetLocalCache 获取本地缓存
func GetLocalCache() *LocalCache {
	once.Do(func() {
		newCache := localcache.New()
		lcache = &LocalCache{Cache: newCache}
	})
	return lcache
}

// SetActiveAlarmCache 设置活动告警缓存
func (l *LocalCache) SetActiveAlarmCache(ruleKey string, value, ttl int64) bool {
	err := retry.Do(func() error {
		success := l.SetWithExpire(fmt.Sprintf(activeTmp, ruleKey), value, ttl)
		if !success {
			return fmt.Errorf("SetActiveAlarmCache cache failed")
		}
		return nil
	}, retry.Attempts(3), retry.RetryIf(func(err error) bool {
		return err != nil
	}))
	return err == nil
}

// CheckActiveAlarmCache 检查活动告警是否存在
func (l *LocalCache) CheckActiveAlarmCache(ruleKey string) bool {
	_, exist := l.Get(fmt.Sprintf(activeTmp, ruleKey))
	return exist
}

// RestoreAlarmCache 删除活动告警缓存
func (l *LocalCache) RestoreAlarmCache(ruleKey string) bool {
	l.Del(fmt.Sprintf(activeTmp, ruleKey))
	return true
}
