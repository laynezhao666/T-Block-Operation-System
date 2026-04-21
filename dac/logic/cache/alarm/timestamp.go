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

	"dac/entity/utils/ttime"
	"gorm.io/gorm"
)

func getDBAlarmsByTimestamp(as []driver.Alarm, timestamp int64) []db.Alarm {
	return getDBAlarms(as, func(a *driver.Alarm) bool {
		return a.Timestamp < timestamp
	})
}

func GetTimestamp(ctx context.Context, controllerID db.IDType, mozuID string) (int64, int64, int64, error) {
	r, err := dac.GetRW().GetOrCreateAlarmTimestampIndex(ctx, controllerID, mozuID)
	if err != nil {
		return 0, 0, 0, err
	}
	return r.CurrentSyncedTimestamp, r.HistoryBeginTimestamp, r.HistorySyncedTimestamp, nil
}

func getAlarmsByTimestamp(controllerID db.IDType, beginTimestamp int64, endTimestamp int64) (driver.AlarmData, error) {
	var (
		result driver.AlarmData
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
	req.Method = driver.MethodGetAlarmByTime
	req.Payload = b

	data, err := dispatcher.Get().DoSyncRequest(req)
	if err != nil {
		return result, fmt.Errorf("fetch controller %v, alarm timestamp %+v error: %w", controllerID, timeInterval, err)
	}

	result, ok = data.(driver.AlarmData)
	if !ok {
		return result, errors.New("type assertion failed")
	}

	return result, nil
}

// FetchByTimestamp 获取指定时间段的告警记录。
// 如果 endTimestamp 为 0，则获取从 beginTimestamp 开始到当前时间的告警记录，更新当前已同步的时间戳。
// 如果 endTimestamp > 0，则获取 [beginTimestamp, endTimestamp) 的告警记录，更新历史已同步的时间戳。
func FetchByTimestamp(ctx context.Context, controllerID db.IDType,
	beginTimestamp int64, endTimestamp int64,
	getControllerFun fetcher.GetControllerFun, _ fetcher.GetArgFun) (int64, error) {
	alarmData, err := getAlarmsByTimestamp(controllerID, beginTimestamp, endTimestamp)
	if err != nil {
		return 0, err
	}
	newTimestamp := alarmData.EndTimestamp

	alarms := getDBAlarmsByTimestamp(alarmData.Alarms, beginTimestamp)

	var (
		c           rt.DoorController
		doorNameMap map[int]string
	)

	fetchHistory := endTimestamp > 0

	if len(alarms) > 0 {
		if getControllerFun != nil {
			c = getControllerFun(ctx)
			doorNameMap = utils.GetDoorNameMap(c.Doors)
		}
	}

	// 获取历史记录时，需要判断插入的记录是否重复。
	// 由于根据时间获取记录的数据中，index 字段无意义，故需要通过控制器id、时间戳、门编号、类型等字段判断是否重复
	if err = dac.GetRW().AddAlarms(ctx, controllerID, alarms, fetchHistory, beginTimestamp, endTimestamp, func(a *db.Alarm) {
		fillAlarm(a, &c, doorNameMap)
	}, func(tx *gorm.DB) error {
		// alarms 可能为空，但需要更新时间戳
		if fetchHistory {
			return dac.UpdateAlarmHistoryTimestampIndex(tx, controllerID, beginTimestamp)
		}
		return dac.UpdateAlarmCurrentTimestampIndex(tx, controllerID, newTimestamp)
	}); err != nil {
		return 0, fmt.Errorf("add controller %v alarms error: %w", controllerID, err)
	}

	return newTimestamp, nil
}
