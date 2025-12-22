package southdevice

import (
	"agent/utils/osal"
	"sync"
	"time"

	"agent/entity/definition"
	"agent/logic/collector/device/model"
	model2 "agent/logic/collector/rtdb/model"
)

var (
	pointsCache = make(map[definition.DeviceGidType]model2.DataPoints)
	pointsMutex sync.RWMutex

	hasCommIDMutex sync.RWMutex
	hasCommIDCache = make(map[definition.DeviceGidType]bool)

	pointsTimer *osal.CronTimer
	commTimer   *osal.CronTimer

	CMax = 6
)

func init() {
	pointsTimer = osal.NewCronTimer(time.Minute*10, func() {
		pointsMutex.Lock()
		defer pointsMutex.Unlock()
		pointsCache = make(map[definition.DeviceGidType]model2.DataPoints)
	})
	pointsTimer.Start()

	commTimer = osal.NewCronTimer(time.Minute*10, func() {
		hasCommIDMutex.Lock()
		defer hasCommIDMutex.Unlock()
		hasCommIDCache = make(map[definition.DeviceGidType]bool)
	})
	commTimer.Start()
}

func getCachedPoints(subDeviceGiD definition.DeviceGidType, pointsInfo model.InstancePointsInfo) model2.DataPoints {
	pointsMutex.RLock()
	if points, ok := pointsCache[subDeviceGiD]; ok {
		pointsMutex.RUnlock()
		return points
	}
	pointsMutex.RUnlock()

	pointsMutex.Lock()
	defer pointsMutex.Unlock()

	if points, ok := pointsCache[subDeviceGiD]; ok {
		return points
	}
	points := pointsInfo.GetDataPoints()
	pointsCache[subDeviceGiD] = points
	return points
}

func hasCommID(subDeviceGiD definition.DeviceGidType) (bool, bool) {
	hasCommIDMutex.RLock()
	defer hasCommIDMutex.RUnlock()

	has, ok := hasCommIDCache[subDeviceGiD]
	return has, ok
}

func setHasCommID(subDeviceGiD definition.DeviceGidType, has bool) {
	hasCommIDMutex.Lock()
	defer hasCommIDMutex.Unlock()

	hasCommIDCache[subDeviceGiD] = has
}
