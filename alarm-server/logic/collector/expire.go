// Package collector report 定时删除Redis中的过期策略valid缓存
package collector

import (
	"context"
	"sync"
	"time"

	"etrpc-go/log"

	"alarm-server/logic/cache"
	"alarm-server/repo/rpc"
	"alarm-server/repo/store"
)

const (
	expireServiceName = "expire"
	// 一小时都没有更新过的key，删除
	InvalidCacheKeyInterval = 3600
)

func expireMozuRecords(mozuId int64) error {
	cacheList, ok := cache.GetLocalCache().GetStrategyKeyCache(mozuId)
	cacheMap, _ := cache.GetLocalCache().GetStrategyCache(mozuId)
	if !ok {
		log.Errorf("expireMozuRecords strategy key list not found")
		return nil
	}
	ridMap, err := store.GetRedisStoreApi().BatchGetRuleRecord(mozuId, cacheList, cacheMap)
	if err != nil {
		log.Errorf("expireMozuRecords BatchGetRuleRecord err: %v", err)
		return err
	}
	delRidFieldMap := make(map[int64][]string) // map[rid] [gid1, gid2, ...]
	var addDelItem = func(rid int64, gid string) {
		if _, ok := delRidFieldMap[rid]; !ok {
			delRidFieldMap[rid] = []string{gid}
		} else {
			delRidFieldMap[rid] = append(delRidFieldMap[rid], gid)
		}
	}
	nowTs := time.Now().Unix()
	for rid, gidMap := range ridMap {
		if gidMap == nil {
			continue
		}
		for gid, v := range gidMap {
			if nowTs-v.EvalTime > InvalidCacheKeyInterval {
				addDelItem(rid, gid)
			}
		}
	}
	store.GetRedisStoreApi().BatchDelRuleRecord(mozuId, delRidFieldMap)
	return nil
}

func expireRecords() {
	mozuList, err := rpc.GetCmdbSvc().GetMozuInfoList()
	if err != nil {
		log.Errorf("expireRecords GetMozuInfoList err: %v", err)
		return
	}
	var wg sync.WaitGroup
	for _, mozuItem := range mozuList {
		wg.Add(1)
		go func(mozuId int64, lwg *sync.WaitGroup) {
			defer lwg.Done()
			success := store.GetRedisStoreApi().TryLock(expireServiceName, mozuId)
			if !success {
				return
			}
			defer func() {
				time.Sleep(10 * time.Second)
				store.GetRedisStoreApi().UnLock(expireServiceName, mozuId)
			}()
			// 上报模组的生效率信息
			err := expireMozuRecords(mozuId)
			if err != nil {
				log.Errorf("expireRecords expireMozuRecords mozuId:%d, err: %v", mozuId, err)
			} else {
				log.Infof("expireRecords mozuId:%v success", mozuId)
			}
		}(int64(mozuItem.MozuId), &wg)
	}
	wg.Wait()
}

// ExpireInvalidRuleRecord 定时删除Redis中的过期策略valid缓存
func ExpireInvalidRuleRecord(_ context.Context) error {
	go expireRecords()
	return nil
}
