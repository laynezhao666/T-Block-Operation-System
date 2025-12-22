package lcache

import (
	"context"
	"sync"
	"time"

	"etrpc-go/log"

	"alarm-manage/conf"
	"alarm-manage/repo/cache"
	"alarm-manage/repo/rpc"
	cmodel "common/entity/model"
)

var (
	cacheAgent *CacheAgent
	once       sync.Once
)

// CacheAgent 缓存agent
type CacheAgent struct {
}

// GetCacheAgent 获取缓存agent
func GetCacheAgent() *CacheAgent {
	once.Do(func() {
		cacheAgent = &CacheAgent{}
	})
	return cacheAgent
}

// 定期同步设备实体信息

// RegularSyncDevice 定期同步全量设备基础信息
func (l *CacheAgent) RegularSyncDevice(ctx context.Context, wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()
	l.checkIfSync(false)
	interval := conf.ServerConf.SyncDeviceCacheConf.RegularSyncDeviceInterval
	TotalCnt := conf.ServerConf.SyncDeviceCacheConf.TotalLoadIntervalCnt
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
				l.checkIfSync(false)
			} else {
				l.checkIfSync(true)
			}
			curCount++
			curCount %= TotalCnt
		}
	}
}

func (l *CacheAgent) checkIfSync(needCheck bool) {
	mozuList, err := rpc.GetCmdbSvc().GetMozuInfoList()
	if err != nil {
		log.Errorf("获取模组信息接口调用失败, err:%v", err)
		return
	}
	for _, mozuItem := range mozuList {
		mozuId, version := mozuItem.MozuId, mozuItem.PublishVersion
		if needCheck {
			if cache.GetLocalCache().NeedUpdateDeviceCache(mozuId, version) {
				l.DoUpdateDeviceCache(mozuId, version)
			}
		} else {
			l.DoUpdateDeviceCache(mozuId, version)
		}
	}
}

// DoUpdateDeviceCache 更新设备实体缓存
func (l *CacheAgent) DoUpdateDeviceCache(mozuId int32, version string) {
	batchSize := conf.ServerConf.SyncDeviceCacheConf.BatchSize
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
			if ok := cache.GetLocalCache().SetDeviceCache(entity.DeviceGid, entity); !ok {
				log.Errorf("DoUpdateDeviceCache set device entity failed, mozuId:%d, deviceGid:%s", mozuId, deviceItem.DeviceGid)
				continue
			}
		}
		totalCnt = int(deviceRsp.Total)
		page++
	}
	if ok := cache.GetLocalCache().SetMozuVersion(mozuId, version); !ok {
		log.Errorf("DoUpdateDeviceCache set mozu version failed, mozuId:%d, version:%s", mozuId, version)
	}
}
