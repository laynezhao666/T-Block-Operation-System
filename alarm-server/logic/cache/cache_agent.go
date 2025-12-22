package cache

import (
	"context"
	"sync"
	"time"

	"etrpc-go/log"

	"alarm-server/conf"
	"alarm-server/repo/rpc"
	cmodel "common/entity/model"
)

// RegularSyncCache 定时同步缓存
func RegularSyncCache(ctx context.Context, wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()
	CheckIfSyncDevice(false)
	CheckIfSyncStrategy(false)
	interval := conf.ServerConf.SyncCacheConfig.DeviceCacheInterval
	TotalCnt := conf.ServerConf.SyncCacheConfig.DeviceTotalIntervalCnt
	if interval == 0 {
		interval = 7200
		TotalCnt = 12
	}
	itv := time.Second * time.Duration(interval)
	tick := time.NewTicker(itv)
	var curCount int32 = 1
	for {
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
			if curCount == 0 {
				CheckIfSyncDevice(false)
				CheckIfSyncStrategy(false)
			} else {
				CheckIfSyncDevice(true)
				CheckIfSyncStrategy(true)
			}
			curCount++
			curCount %= TotalCnt
		}
	}
}

// CheckIfSyncDevice 检查是否需要同步设备缓存
func CheckIfSyncDevice(needCheck bool) {
	mozuList, err := rpc.GetCmdbSvc().GetMozuInfoList()
	if err != nil {
		log.Errorf("checkIfSync get mozu info failed, err:%v", err)
		return
	}
	for _, mozuItem := range mozuList {
		mozuId, version := int64(mozuItem.MozuId), mozuItem.PublishVersion
		if needCheck {
			if GetLocalCache().NeedUpdateDeviceCache(mozuId, version) {
				DoUpdateDeviceCache(mozuId, version)
			}
		} else {
			DoUpdateDeviceCache(mozuId, version)
		}
	}
}

// DoUpdateDeviceCache 更新设备实体缓存
func DoUpdateDeviceCache(mozuId int64, version string) {
	batchSize := conf.ServerConf.SyncCacheConfig.DeviceCacheBatchSize
	if batchSize == 0 {
		batchSize = 8000
	}
	totalCnt := 1
	page := 1
	for (page-1)*int(batchSize) < totalCnt {
		deviceRsp, err := rpc.GetCmdbSvc().GetDeviceEntity(mozuId, page, int(batchSize))
		if err != nil {
			log.Errorf("DoUpdateDeviceCache get device entity failed, mozuId:%d, err:%v", mozuId, err)
			return
		}
		for _, deviceItem := range deviceRsp.List {
			entity := cmodel.DeviceEntity{
				DeviceGid:         deviceItem.DeviceGid,
				DeviceNumber:      deviceItem.DeviceNumber,
				DeviceName:        deviceItem.DeviceName,
				MozuId:            deviceItem.MozuId,
				MozuName:          deviceItem.MozuName,
				IdcArea:           deviceItem.IdcArea,
				FuncRoom:          deviceItem.FuncRoom,
				DeviceTypeEn:      deviceItem.DeviceTypeEn,
				DeviceTypeZh:      deviceItem.DeviceTypeZh,
				ApplicationTypeEn: deviceItem.ApplicationTypeEn,
				ApplicationTypeZh: deviceItem.ApplicationTypeZh,
			}
			if ok := GetLocalCache().SetDeviceCache(entity.DeviceGid, entity); !ok {
				log.Errorf("DoUpdateDeviceCache set device entity failed, mozuId:%d, deviceGid:%s", mozuId, deviceItem.DeviceGid)
				continue
			}
		}
		totalCnt = int(deviceRsp.Total)
		page++
	}
	if ok := GetLocalCache().SetMozuVersion(mozuId, version); !ok {
		log.Errorf("DoUpdateDeviceCache set mozu version failed, mozuId:%d, version:%s", mozuId, version)
	}
}
