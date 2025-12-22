package cache

import (
	"sync"

	"trpc.group/trpc-go/trpc-database/localcache"
)

var (
	lcache *LocalCache
	once   sync.Once
)

// LocalCache 本地缓存
type LocalCache struct {
	localcache.Cache
}

// GetLocalCache 获取本地缓存
func GetLocalCache() *LocalCache {
	once.Do(func() {
		lcache = &LocalCache{
			localcache.New(),
		}
	})
	return lcache
}
