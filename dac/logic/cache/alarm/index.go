// Package alarm 提供门禁告警的缓存和增量拉取功能。
package alarm

import (
	"context"
	"errors"
	"fmt"

	"dac/entity/model/db"
	"dac/entity/model/driver"
	"dac/entity/model/rt"
	"dac/entity/utils"
	"dac/logic/cache/fetcher"
	"dac/logic/collect/dispatcher"
	"dac/repo/dac"

	"gorm.io/gorm"
)

// getDBAlarmsByIndex 过滤索引之前的告警并转换为数据库模型
func getDBAlarmsByIndex(as []driver.Alarm, currentIndex int) []db.Alarm {
	return getDBAlarms(as, func(a *driver.Alarm) bool {
		return a.Index <= currentIndex
	})
}

// GetIndex 获取告警索引记录（当前索引和最新索引）
func GetIndex(ctx context.Context, controllerID db.IDType) (int, int, error) {
	r, err := dac.GetRW().GetOrCreateAlarmIndex(ctx, controllerID)
	if err != nil {
		return 0, 0, err
	}
	return r.Index, r.Last, nil
}

// getAlarmsByIndex 从控制器获取指定索引的告警数据
func getAlarmsByIndex(controllerID db.IDType, index int) (driver.AlarmData, error) {
	var (
		result driver.AlarmData
		ok     bool
	)
	b, err := driver.Marshal(index)
	if err != nil {
		return result, err
	}

	req := new(db.Request)
	req.ControllerID = controllerID
	req.Method = driver.MethodGetAlarm
	req.Payload = b

	data, err := dispatcher.Get().DoSyncRequest(req)
	if err != nil {
		return result, fmt.Errorf("fetch controller %v, alarm index %v error: %w", controllerID, index, err)
	}

	result, ok = data.(driver.AlarmData)
	if !ok {
		return result, errors.New("type assertion failed")
	}

	return result, nil
}

// Fetch 按索引增量拉取告警数据并写入数据库。
// 当 index == last 时先检查是否有新记录，若 last 变小则说明控制器被重置。
func Fetch(ctx context.Context, controllerID db.IDType, index, last int,
	getControllerFun fetcher.GetControllerFun,
	_ fetcher.GetArgFun,
) (int, int, error) {
	requestNewLast := index == last && last > 0
	if requestNewLast {
		alarmData, err := getAlarmsByIndex(controllerID, 0)
		if err != nil {
			return 0, 0, err
		}
		// 门禁记录已获取到最后一条
		if alarmData.Last == last {
			return index, last, nil
		} else if alarmData.Last < last {
			// 门禁记录变少，可能更换门禁，需要重新获取
			return 0, 0, nil
		}
	}

	alarmData, err := getAlarmsByIndex(controllerID, index)
	if err != nil {
		return 0, 0, err
	}

	// 只需要取大于当前索引的数据
	alarms := getDBAlarmsByIndex(alarmData.Alarms, index)
	newIndex := alarmData.Offset
	newLast := alarmData.Last

	var (
		c           rt.DoorController
		doorNameMap map[int]string
	)

	if len(alarms) > 0 {
		if getControllerFun != nil {
			c = getControllerFun(ctx)
			doorNameMap = utils.GetDoorNameMap(c.Doors)
		}
	}

	if err = dac.GetRW().AddAlarms(ctx, controllerID, alarms, false, 0, 0, func(a *db.Alarm) {
		fillAlarm(a, &c, doorNameMap)
	}, func(tx *gorm.DB) error {
		return dac.UpdateAlarmIndex(tx, controllerID, newIndex, newLast)
	}); err != nil {
		return 0, 0, fmt.Errorf("add controller %v alarms error: %w", controllerID, err)
	}

	return newIndex, newLast, nil
}
