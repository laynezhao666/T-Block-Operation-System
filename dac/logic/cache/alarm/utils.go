// Package alarm 提供门禁告警的缓存和增量拉取功能。
package alarm

import (
	"dac/entity/model/db"
	"dac/entity/model/driver"
	"dac/entity/model/rt"
)

// getDBAlarms 将驱动层告警数据转换为数据库模型，支持自定义过滤
func getDBAlarms(as []driver.Alarm, filter func(*driver.Alarm) bool) []db.Alarm {
	alarms := make([]db.Alarm, 0, len(as))
	for i := range as {
		if filter(&as[i]) {
			continue
		}

		alarms = append(alarms, db.Alarm{
			Index:       as[i].Index,
			Timestamp:   as[i].Timestamp,
			DoorNumber:  int(as[i].DoorNumber),
			Type:        int(as[i].Type),
			State:       int(as[i].State),
			Description: as[i].Description,
		})
	}
	return alarms
}

// fillAlarm 填充告警的控制器名称、门名称和模组ID
func fillAlarm(a *db.Alarm, c *rt.DoorController, doorNameMap map[int]string) {
	a.ControllerName = c.Name
	a.DoorName = doorNameMap[a.DoorNumber]
	a.MozuID = c.MozuID
}
