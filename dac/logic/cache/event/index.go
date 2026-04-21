// Package event 提供门禁事件的缓存和增量拉取功能。
package event

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

// getDBEventsByIndex 过滤索引之前的事件并转换为数据库模型
func getDBEventsByIndex(es []driver.Event, currentIndex int) []db.Event {
	return getDBEvents(es, func(e *driver.Event) bool {
		return e.Index <= currentIndex
	})
}

// GetIndex 获取事件索引记录（当前索引和最新索引）
func GetIndex(ctx context.Context, controllerID db.IDType) (int, int, error) {
	r, err := dac.GetRW().GetOrCreateEventIndex(ctx, controllerID)
	if err != nil {
		return 0, 0, err
	}
	return r.Index, r.Last, nil
}

// getEventsByIndex 从控制器获取指定索引的事件数据
func getEventsByIndex(controllerID db.IDType, index int) (driver.EventData, error) {
	var (
		result driver.EventData
		ok     bool
	)
	b, err := driver.Marshal(index)
	if err != nil {
		return result, err
	}

	req := new(db.Request)
	req.ControllerID = controllerID
	req.Method = driver.MethodGetEvent
	req.Payload = b

	data, err := dispatcher.Get().DoSyncRequest(req)
	if err != nil {
		return result, fmt.Errorf("fetch controller %v, event index %v error: %w", controllerID, index, err)
	}

	result, ok = data.(driver.EventData)
	if !ok {
		return result, errors.New("type assertion failed")
	}

	return result, nil
}

// Fetch 按索引增量拉取事件数据并写入数据库。
// 当 index == last 时先检查是否有新记录，若 last 变小则说明控制器被重置。
func Fetch(ctx context.Context, controllerID db.IDType, index, last int,
	getControllerFun fetcher.GetControllerFun,
	getArgFun fetcher.GetArgFun,
) (int, int, error) {
	requestNewLast := index == last && last > 0
	if requestNewLast {
		eventData, err := getEventsByIndex(controllerID, 0)
		if err != nil {
			return 0, 0, err
		}
		// 门禁记录已获取到最后一条
		if eventData.Last == last {
			return index, last, nil
		} else if eventData.Last < last {
			// 门禁记录变少，可能更换门禁，需要重新获取
			return 0, 0, nil
		}
	}

	eventData, err := getEventsByIndex(controllerID, index)
	if err != nil {
		return 0, 0, err
	}

	// 只需要取大于当前索引的数据
	events := getDBEventsByIndex(eventData.Events, index)
	newIndex := eventData.Offset
	newLast := eventData.Last

	var (
		c            rt.DoorController
		doorNameMap  map[int]string
		arg          interface{} = nil
		cardStaffMap map[string]db.Staff
	)

	if len(events) > 0 {
		if getControllerFun != nil {
			c = getControllerFun(ctx)
			doorNameMap = utils.GetDoorNameMap(c.Doors)
		}
		if getArgFun != nil {
			arg = getArgFun(ctx)
			cardStaffMap, _ = arg.(map[string]db.Staff)
		}
	}

	if err = dac.GetRW().AddEvents(ctx, controllerID, events, false, 0, 0, func(e *db.Event) {
		fillEvent(e, &c, doorNameMap, cardStaffMap)
	}, func(tx *gorm.DB) error {
		return dac.UpdateEventIndex(tx, controllerID, newIndex, newLast)
	}); err != nil {
		return 0, 0, fmt.Errorf("add controller %v events error: %w", controllerID, err)
	}

	return newIndex, newLast, nil
}
