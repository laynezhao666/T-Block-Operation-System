// Package cache 用于缓存部分数据库数据
package cache

import (
	"fmt"
	"runtime"
	"sync"
	"time"

	"trpc.group/trpc-go/trpc-database/localcache"
	"trpc.group/trpc-go/trpc-go/log"
)

// IDataCache 数据缓存器
type IDataCache[T any] interface {
	GetData(mozuId int32) (T, bool)
}

// NewMozuDataCache 注册一个数据缓存对象
func NewMozuDataCache[T any](name string, loader IDataLoader[T], expireSec int64) IDataCache[T] {
	return &dataCacheImpl[T]{
		name:   name,
		loader: loader,
		ttlSec: expireSec,
		data:   localcache.New(localcache.WithSettingTimeout(time.Minute)),
	}
}

type dataCacheImpl[T any] struct {
	name   string           // 缓存名称
	loader IDataLoader[T]   // 数据加载器
	data   localcache.Cache // 内存缓存的数据,按模组缓存
	ttlSec int64            // 缓存过期时间
	curVer sync.Map         // 当前各模组数据版本
	lock   sync.Map         // 每个模组对应的锁
	order  int32            // 加载顺序,数值越低,越先加载
}

func (t *dataCacheImpl[T]) GetData(mozuId int32) (T, bool) {
	curVer, ok := mozuVerMap.Load(mozuId)
	// 模组ID不存在, 直接返回空
	if !ok {
		return *new(T), false
	}
	lastVer, ok1 := t.curVer.Load(mozuId)
	res, ok2 := t.data.Get(fmt.Sprint(mozuId))
	// 上次版本不存在或者版本发生变化, 后者数据已经过期,重新缓存数据
	if !ok1 || curVer.(string) != lastVer.(string) || !ok2 || res == nil {
		t.DoCache(mozuId, curVer.(string))
		res, ok2 = t.data.Get(fmt.Sprint(mozuId))
	} else {
		// 重置时间
		t.data.SetWithExpire(fmt.Sprint(mozuId), res, t.ttlSec)
	}
	return res.(T), ok2
}

func (t *dataCacheImpl[T]) DoCache(mozuId int32, curVer string) {
	// 加锁进行并发控制
	lockVal, _ := t.lock.LoadOrStore(mozuId, &sync.Mutex{})
	lock := lockVal.(*sync.Mutex)
	lock.Lock()
	defer lock.Unlock()
	// 二次判断是否需要重新加载
	lastVer, ok1 := t.curVer.Load(mozuId)
	data, ok2 := t.data.Get(fmt.Sprint(mozuId))
	if ok1 && lastVer.(string) == curVer && ok2 && data != nil {
		return
	}
	mozuData, err := t.loader.Load(mozuId)
	if err != nil {
		log.Errorf("cache:[%s] refresh mozu:[%d] cache for ver:[%s] fail, err: %v", t.name, mozuId, curVer, err)
		return
	}
	t.data.SetWithExpire(fmt.Sprint(mozuId), mozuData, t.ttlSec)
	t.curVer.Store(mozuId, curVer)
	runtime.GC()
	log.Infof("cache:[%s] refresh mozu:[%d] cache for ver:[%s] success", t.name, mozuId, curVer)
}
