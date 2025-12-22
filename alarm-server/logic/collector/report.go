// Package collector report 生效率上报智研
package collector

import (
	"context"
	"fmt"
	"sync"
	"time"

	"etrpc-go/log"

	"alarm-server/logic/api/strategy"
	"alarm-server/logic/cache"
	"alarm-server/repo/ckafka"
	"alarm-server/repo/rpc"
	"alarm-server/repo/store"
	"alarm-server/utils/modcall"
)

const (
	reportServiceName = "report"
)

func reportMozuEfficiency(mozuId int64) error {
	cacheMap, strategyOk := cache.GetLocalCache().GetStrategyCache(mozuId)
	cntMap, cntOk := cache.GetLocalCache().GetMozuStrategyCnt(mozuId)
	if !cntOk || !strategyOk {
		log.Errorf("reportMozuEfficiency get strategy cache is empty, mozuId:%d", mozuId)
		return fmt.Errorf("reportMozuEfficiency get strategy cache is empty, mozuId:%d", mozuId)
	}
	reportFunc := func(service string, list []int64) {
		recordMap, metricMap, err :=
			strategy.FilterRuleRecord(mozuId, list, cacheMap, cntMap, -1)
		if err != nil {
			log.Errorf("reportMozuEfficiency failed, service:%s, mozuId:%d, err:%v", service, mozuId, err)
			return
		}
		validCnt := cntMap["all"] - metricMap["all"][1]
		modcall.RecordDistinctAnalyzeCnt(service, int(mozuId), int(validCnt), int(metricMap["all"][0]), int(cntMap["all"]))
		ckafka.GetCkafka().BatchSendAdminRuleMsg(recordMap)
	}
	list := []int64{}
	for rid := range cacheMap {
		list = append(list, rid)
	}
	reportFunc("TotalRT", list)
	return nil
}

// CalEfficiencyMetric 定时汇总生效率指标
func CalEfficiencyMetric() {
	mozuList, err := rpc.GetCmdbSvc().GetMozuInfoList()
	if err != nil {
		log.Errorf("CalEfficiencyMetric GetMozuInfoList err: %v", err)
		return
	}
	var wg sync.WaitGroup
	for _, mozuItem := range mozuList {
		wg.Add(1)
		go func(mozuId int64, lwg *sync.WaitGroup) {
			defer lwg.Done()
			success := store.GetRedisStoreApi().TryLock(reportServiceName, mozuId)
			if !success {
				return
			}
			defer func() {
				time.Sleep(10 * time.Second)
				store.GetRedisStoreApi().UnLock(reportServiceName, mozuId)
			}()
			// 上报模组的生效率信息
			err := reportMozuEfficiency(mozuId)
			if err != nil {
				log.Errorf("CalEfficiencyMetric reportMozuEfficiency mozuId:%d, err: %v", mozuId, err)
			} else {
				log.Infof("CalEfficiencyMetric mozuId:%v success", mozuId)
			}
		}(int64(mozuItem.MozuId), &wg)
	}
	wg.Wait()
}

// ReportValidEfficiency 定时汇总生效率指标
func ReportValidEfficiency(_ context.Context) error {
	go CalEfficiencyMetric()
	return nil
}
