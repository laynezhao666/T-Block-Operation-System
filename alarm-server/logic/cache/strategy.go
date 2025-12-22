// Package cache 缓存策略信息
package cache

import (
	"etrpc-go/log"

	"trpc.group/trpc-go/trpc-go"

	"alarm-server/conf"
	"alarm-server/entity/model"
	"alarm-server/repo/dao/strategy"
	"alarm-server/repo/rpc"
	"alarm-server/utils/common"
	cmodel "common/entity/model"
)

// CheckIfSyncStrategy 检查是否需要同步策略缓存
func CheckIfSyncStrategy(needCheck bool) {
	mozuList, err := rpc.GetCmdbSvc().GetMozuInfoList()
	if err != nil {
		log.Errorf("checkIfSync get strategy metrics failed, err:%v", err)
		return
	}
	for _, mozuItem := range mozuList {
		mozuId, alarmVersion := int64(mozuItem.MozuId), mozuItem.AlarmVersion
		if needCheck {
			if GetLocalCache().NeedUpdateStrategy(mozuId, alarmVersion) {
				DoUpdateStrategyCache(mozuId, alarmVersion)
			}
		} else {
			DoUpdateStrategyCache(mozuId, alarmVersion)
		}
	}

}

// DoUpdateStrategyCache 更新策略缓存
func DoUpdateStrategyCache(mozuId int64, version string) {
	batchSize := conf.ServerConf.SyncCacheConfig.StrategyCacheBatchSize
	if batchSize == 0 {
		batchSize = 30000
	}
	list := []cmodel.AlarmStrategy{}
	page := 1
	totalCnt := 1
	cntMap := map[string]int32{
		"all": 0,
		"L0":  0,
		"L1":  0,
		"L2":  0,
		"L3":  0,
		"L4":  0,
	}
	for (page-1)*int(batchSize) < totalCnt {
		batchList, cnt, err := strategy.NewStrategyDao().GetStrategyList(trpc.BackgroundContext(), &strategy.StrategyFilter{
			MozuId:  mozuId,
			RidType: []int64{0, 1},
			Page:    int64(page),
			Size:    int64(batchSize),
		})
		if err != nil {
			log.Errorf("DoUpdateStrategyCache get strategy list failed, mozuId:%d, err:%v", mozuId, err)
			return
		}
		list = append(list, batchList...)
		totalCnt = int(cnt)
		page++
	}
	strategyMap := map[int64]*common.OrderedMap[string, model.StrategyCacheData]{}
	for _, item := range list {
		if _, ok := strategyMap[item.Rid]; !ok {
			strategyMap[item.Rid] = common.NewOrderedMap[string, model.StrategyCacheData]()
		}
		strategyMap[item.Rid].Set(item.DeviceGid, model.StrategyCacheData{
			ID:                   item.Id,
			DeviceGid:            item.DeviceGid,
			Rid:                  item.Rid,
			RidVersion:           item.RidVersion,
			RidType:              item.RidType,
			MozuId:               item.MozuId,
			AlarmName:            item.AlarmName,
			AlarmExpression:      item.AlarmExpression,
			AlarmExpressionStr:   item.AlarmExpressionStr,
			RestoreExpression:    item.RestoreExpression,
			RestoreExpressionStr: item.RestoreExpressionStr,
			ExpressionMap:        item.GetExprMap(),
			AlarmLevel:           item.AlarmLevel,
			ContentTemplate:      item.ContentTemplate,
			Owner:                item.Owner,
			UpdateAt:             item.UpdateAt,
			CreateAt:             item.CreateAt,
		})
		if _, ok := cntMap[item.AlarmLevel]; ok {
			cntMap[item.AlarmLevel] = cntMap[item.AlarmLevel] + 1
		}
		cntMap["all"] = cntMap["all"] + 1
	}
	success := GetLocalCache().SetStrategyCache(mozuId, version, cntMap, strategyMap)
	if !success {
		log.Errorf("DoUpdateStrategyCache set strategy cache failed, mozuId:%d, version:%s", mozuId, version)
	}
}
