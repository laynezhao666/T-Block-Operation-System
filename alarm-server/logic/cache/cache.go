// Package cache 本地缓存
package cache

import (
	"fmt"
	"sync"

	"etrpc-go/log"

	"github.com/avast/retry-go"
	"github.com/samber/lo"
	"trpc.group/trpc-go/trpc-database/localcache"

	"alarm-server/entity/model"
	"alarm-server/utils/common"
	cmodel "common/entity/model"
)

var (
	lcache *LocalCache
	once   sync.Once
)

const (
	mozuVersionKeyTmp       = "version:%d" // map[mozuId] version
	deviceCacheKeyTmp       = "attr:%s"    // map[gid]
	strategyCntTemplate     = "cnt:%d"
	strategyVersionTemplate = "rule:%d"
	strategyCacheTemplate   = "cache:%d"
	strategyKeyTemplate     = "key:%d"
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

// ----------------------------------------------------策略缓存---------------------------------------------------------//

// NeedUpdateStrategy 是否需要更新策略
func (l *LocalCache) NeedUpdateStrategy(mozuId int64, version string) bool {
	curAlarmVersion, ok := l.Get(fmt.Sprintf(strategyVersionTemplate, mozuId))
	if !ok {
		return true
	}
	if curAlarmVersion.(string) != version {
		return true
	}
	return false
}

// SetStrategyCache 设置策略状态
// @param cnt 策略数量
// @param version 策略版本
func (l *LocalCache) SetStrategyCache(mozuId int64, version string, cntMap map[string]int32,
	strategyMap map[int64]*common.OrderedMap[string, model.StrategyCacheData]) bool {
	keyList := lo.Keys(strategyMap)
	err := retry.Do(func() error {
		successStrategy := l.SetWithExpire(fmt.Sprintf(strategyCacheTemplate, mozuId), strategyMap, 3*24*60*60)
		successRid := l.SetWithExpire(fmt.Sprintf(strategyKeyTemplate, mozuId), keyList, 3*24*60*60)
		successCnt := l.SetWithExpire(fmt.Sprintf(strategyCntTemplate, mozuId), cntMap, 3*24*60*60)
		successVersion := l.SetWithExpire(fmt.Sprintf(strategyVersionTemplate, mozuId), version, 3*24*60*60)
		if !successStrategy || !successRid || !successCnt || !successVersion {
			return fmt.Errorf("SetStrategyCache cache failed, mozuId:%d", mozuId)
		}
		return nil
	}, retry.Attempts(3), retry.RetryIf(func(err error) bool {
		return err != nil
	}))
	return err == nil
}

// GetMozuStrategyCnt GetMozuStrategyCnt
func (l *LocalCache) GetMozuStrategyCnt(mozuId int64) (map[string]int32, bool) {
	ret, ok := l.Get(fmt.Sprintf(strategyCntTemplate, mozuId))
	if !ok {
		log.Errorf("GetMozuStrategyCnt get strategy cnt failed, mozuId:%d", mozuId)
		return nil, false
	}
	return ret.(map[string]int32), true
}

// GetStrategyCache GetStrategyCache
func (l *LocalCache) GetStrategyCache(mozuId int64) (map[int64]*common.OrderedMap[string, model.StrategyCacheData], bool) {
	ret, ok := l.Get(fmt.Sprintf(strategyCacheTemplate, mozuId))
	if !ok {
		return nil, false
	}
	return ret.(map[int64]*common.OrderedMap[string, model.StrategyCacheData]), true
}

// GetStrategyKeyCache GetStrategyKeyCache
func (l *LocalCache) GetStrategyKeyCache(mozuId int64) ([]int64, bool) {
	ret, ok := l.Get(fmt.Sprintf(strategyKeyTemplate, mozuId))
	if !ok {
		return nil, false
	}
	return ret.([]int64), true
}

// -------------------------------------------------设备缓存----------------------------------------------------------//

// NeedUpdateDeviceCache 是否需要更新设备缓存
func (l *LocalCache) NeedUpdateDeviceCache(mozuId int64, version string) bool {
	curVersion, ok := l.Get(fmt.Sprintf(mozuVersionKeyTmp, mozuId))
	if !ok {
		return true
	}
	if curVersion.(string) != version {
		return true
	}
	return false
}

// SetMozuVersion SetMozuVersion
func (l *LocalCache) SetMozuVersion(mozuId int64, version string) bool {
	err := retry.Do(func() error {
		success := l.SetWithExpire(fmt.Sprintf(mozuVersionKeyTmp, mozuId), version, 3*24*60*60)
		if !success {
			return fmt.Errorf("SetMozuVersion cache failed")
		}
		return nil
	}, retry.Attempts(3), retry.RetryIf(func(err error) bool {
		return err != nil
	}))
	return err == nil
}

// SetDeviceCache SetDeviceCache
func (l *LocalCache) SetDeviceCache(gid string, entity cmodel.DeviceEntity) bool {
	err := retry.Do(func() error {
		success := l.SetWithExpire(fmt.Sprintf(deviceCacheKeyTmp, gid), entity, 3*24*60*60)
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
	ret, ok := l.Get(fmt.Sprintf(deviceCacheKeyTmp, gid))
	if !ok {
		return nil, false
	}
	res := ret.(cmodel.DeviceEntity)
	return &res, true
}
