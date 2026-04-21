// Package event 提供门禁事件的缓存和基于时间戳的增量拉取功能。
package event

import (
	"context"
	"errors"
	"fmt"

	"dac/entity/config"
	"dac/entity/model/db"
	"dac/entity/model/driver"
	"dac/entity/model/rt"
	"dac/entity/utils"
	"dac/logic/cache/fetcher"
	"dac/logic/collect/dispatcher"
	"dac/repo/dac"

	"dac/entity/utils/ttime"
	"gorm.io/gorm"
)

// getDBEventsByTimestamp 过滤时间戳之前的事件并转换为数据库模型
func getDBEventsByTimestamp(as []driver.Event, timestamp int64) []db.Event {
	return getDBEvents(as, func(e *driver.Event) bool {
		return e.Timestamp < timestamp
	})
}

// GetTimestamp 获取事件时间戳索引（当前同步、历史开始、历史同步时间戳）
func GetTimestamp(ctx context.Context, controllerID db.IDType,
	mozuID string,
) (int64, int64, int64, error) {
	r, err := dac.GetRW().GetOrCreateEventTimestampIndex(ctx, controllerID, mozuID)
	if err != nil {
		return 0, 0, 0, err
	}
	return r.CurrentSyncedTimestamp, r.HistoryBeginTimestamp, r.HistorySyncedTimestamp, nil
}

// getEventsByTimestamp 从控制器获取指定时间段的事件数据
func getEventsByTimestamp(controllerID db.IDType,
	beginTimestamp int64, endTimestamp int64,
) (driver.EventData, error) {
	var (
		result driver.EventData
		ok     bool
	)
	timeInterval := driver.TimeInterval{
		BeginTimestamp: beginTimestamp,
		EndTimestamp:   endTimestamp,
	}
	if endTimestamp <= 0 {
		timeInterval.EndTimestamp = ttime.GetNowUTC().Unix()
	}

	b, err := driver.Marshal(timeInterval)
	if err != nil {
		return result, err
	}

	req := new(db.Request)
	req.ControllerID = controllerID
	req.Method = driver.MethodGetEventByTime
	req.Payload = b

	data, err := dispatcher.Get().DoSyncRequest(req)
	if err != nil {
		return result, fmt.Errorf("fetch controller %v, event timestamp %+v error: %w", controllerID, timeInterval, err)
	}

	result, ok = data.(driver.EventData)
	if !ok {
		return result, errors.New("type assertion failed")
	}

	return result, nil
}

// FetchByTimestamp 获取指定时间段的事件记录。
// 如果 endTimestamp 为 0，则获取从 beginTimestamp 开始到当前时间的事件记录，更新当前已同步的时间戳。
// 如果 endTimestamp > 0，则获取 [beginTimestamp, endTimestamp) 的事件记录，更新历史已同步的时间戳。
func FetchByTimestamp(ctx context.Context, controllerID db.IDType,
	beginTimestamp int64, endTimestamp int64,
	getControllerFun fetcher.GetControllerFun,
	getArgFun fetcher.GetArgFun,
) (int64, error) {
	eventData, err := getEventsByTimestamp(controllerID, beginTimestamp, endTimestamp)
	if err != nil {
		return 0, err
	}
	newTimestamp := eventData.EndTimestamp

	events := getDBEventsByTimestamp(eventData.Events, beginTimestamp)

	var (
		c            rt.DoorController
		doorNameMap  map[int]string
		arg          interface{} = nil
		cardStaffMap map[string]db.Staff
	)

	fetchHistory := endTimestamp > 0

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

	// 获取历史记录时，需要判断插入的记录是否重复。
	// 由于根据时间获取记录的数据中，index 字段无意义，故需要通过控制器id、时间戳、门编号、类型等字段判断是否重复
	if err = dac.GetRW().AddEvents(ctx, controllerID, events, fetchHistory, beginTimestamp, endTimestamp, func(e *db.Event) {
		fillEvent(e, &c, doorNameMap, cardStaffMap)
	}, func(tx *gorm.DB) error {
		// events 可能为空，但需要更新时间戳
		config.Log.Infof("update event history timestamp index"+
			" controller_id: %v, begin_timestamp: %v, end_timestamp: %v"+
			", fetch history: %v", controllerID, beginTimestamp, endTimestamp, fetchHistory)
		if fetchHistory {
			return dac.UpdateEventHistoryTimestampIndex(tx, controllerID, beginTimestamp)
		}
		return dac.UpdateEventCurrentTimestampIndex(tx, controllerID, newTimestamp)
	}); err != nil {
		return 0, fmt.Errorf("add controller %v events error: %w", controllerID, err)
	}

	return newTimestamp, nil
}
