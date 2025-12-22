// Package cache 本地缓存
package cache

import (
	"fmt"
	"sync"

	"github.com/avast/retry-go"
	"github.com/samber/lo"
	"trpc.group/trpc-go/trpc-database/localcache"

	cmodel "common/entity/model"
)

var (
	lcache *LocalCache
	cOnce  sync.Once
)

const (
	// mozuId 级别
	mozuVersionKeyTmp = "version:%d" // map[mozuId] version
	deviceCacheKeyTmp = "attr:%s"    // map[gid]
	mozuSyncCloudTmp  = "sync_cloud"
)

// LocalCache 本地缓存
type LocalCache struct {
	// 本地缓存，localcache设置过期时间
	cache localcache.Cache

	// 云端缓存，map[mozuId]struct{}{}，不设置过期时间
	cloudMutex sync.RWMutex
	cloudCache map[int32]struct{}
}

// GetLocalCache 获取本地缓存
func GetLocalCache() *LocalCache {
	cOnce.Do(func() {
		newCache := localcache.New()
		lcache = &LocalCache{
			cache:      newCache,
			cloudCache: make(map[int32]struct{}),
		}
	})
	return lcache
}

/*-----------------------------------模组Version-----------------------------------------*/

// NeedUpdateDeviceCache 是否需要更新设备缓存
func (l *LocalCache) NeedUpdateDeviceCache(mozuId int32, version string) bool {
	curVersion, ok := l.cache.Get(fmt.Sprintf(mozuVersionKeyTmp, mozuId))
	if !ok {
		return true
	}
	if curVersion.(string) != version {
		return true
	}
	return false
}

// SetMozuVersion SetMozuVersion
func (l *LocalCache) SetMozuVersion(mozuId int32, version string) bool {
	err := retry.Do(func() error {
		success := l.cache.SetWithExpire(fmt.Sprintf(mozuVersionKeyTmp, mozuId), version, 3*24*60*60)
		if !success {
			return fmt.Errorf("SetMozuVersion cache failed")
		}
		return nil
	}, retry.Attempts(3), retry.RetryIf(func(err error) bool {
		return err != nil
	}))
	return err == nil
}

/*-----------------------------------设备缓存-----------------------------------------*/

// SetDeviceCache SetDeviceCache
func (l *LocalCache) SetDeviceCache(gid string, entity cmodel.DeviceEntity) bool {
	err := retry.Do(func() error {
		success := l.cache.SetWithExpire(fmt.Sprintf(deviceCacheKeyTmp, gid), entity, 3*24*60*60)
		if !success {
			return fmt.Errorf("SetMozuVersion cache failed")
		}
		return nil
	}, retry.Attempts(3), retry.RetryIf(func(err error) bool {
		return err != nil
	}))
	return err == nil
}

// GetDeviceCache GetDeviceCache
func (l *LocalCache) GetDeviceCache(gid string) (*cmodel.DeviceEntity, bool) {
	ret, ok := l.cache.Get(fmt.Sprintf(deviceCacheKeyTmp, gid))
	if !ok {
		return nil, false
	}
	res := ret.(cmodel.DeviceEntity)
	return &res, true
}

/*-----------------------------------模组是否推送云端配置-----------------------------------------*/

// SetCloudAccessCache 设置发往云端的模组配置
func (l *LocalCache) SetCloudAccessCache(mozuList []int32) bool {
	if mozuList == nil {
		return false
	}
	mozuMap := lo.SliceToMap(mozuList, func(item int32) (int32, struct{}) {
		return item, struct{}{}
	})
	l.cloudMutex.Lock()
	l.cloudCache = mozuMap
	l.cloudMutex.Unlock()
	return true
}

// CheckIfSyncCloud 检查模组是否同步云端
func (l *LocalCache) CheckIfSyncCloud(mozuId int32) bool {
	l.cloudMutex.RLock()
	defer l.cloudMutex.RUnlock()
	if l.cloudCache == nil {
		return false
	}
	_, ok := l.cloudCache[mozuId]
	return ok
}
