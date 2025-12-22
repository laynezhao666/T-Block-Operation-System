package southdevice

import (
	"context"
	"agent/entity/definition"
	"sync"
	"time"
)

const (
	refreshCacheTime = time.Minute * 10
)

var (
	commDevicesCache = newDeviceCache(context.TODO())
)

type deviceSet map[definition.DeviceGidType]struct{}
type deviceCacheType struct {
	m     deviceSet
	mutex sync.RWMutex
}

func newDeviceCache(ctx context.Context) *deviceCacheType {
	d := new(deviceCacheType)
	d.m = make(deviceSet)
	go d.refreshLoop(ctx)
	return d
}

func getCommDeviceCache() *deviceCacheType {
	return commDevicesCache
}

func (d *deviceCacheType) refresh(ctx context.Context) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	d.m = make(deviceSet)
}

func (d *deviceCacheType) refreshLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(refreshCacheTime):
			d.refresh(ctx)
		}
	}
}

// AddDevices 添加设备
func (d *deviceCacheType) AddDevices(id definition.DeviceGidType) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	d.m[id] = struct{}{}
}

// HasCommePoint 判断是否是通用设备
func (d *deviceCacheType) HasCommePoint(id definition.DeviceGidType) bool {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	_, ok := d.m[id]
	return ok
}
