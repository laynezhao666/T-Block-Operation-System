// Package event 提供门禁事件的缓存和增量拉取功能。
package event

import (
	"dac/entity/consts"
	"dac/entity/model/db"
	"dac/entity/model/driver"
	"dac/entity/model/rt"
)

// getDBEvents 将驱动层事件数据转换为数据库模型，支持自定义过滤
func getDBEvents(es []driver.Event, filter func(*driver.Event) bool) []db.Event {
	events := make([]db.Event, 0, len(es))
	for i := range es {
		if filter(&es[i]) {
			continue
		}

		events = append(events, db.Event{
			Index:       es[i].Index,
			Timestamp:   es[i].Timestamp,
			CardNumber:  es[i].CardNumber,
			Username:    es[i].Username,
			DoorNumber:  int(es[i].DoorNumber),
			Direction:   int(es[i].Direction),
			Type:        int(es[i].Type),
			Description: es[i].Description,
		})
	}
	return events
}

// fillEvent 填充事件的控制器名称、门名称、模组ID和人员信息
func fillEvent(
	e *db.Event, c *rt.DoorController,
	doorNameMap map[int]string,
	cardStaffMap map[string]db.Staff,
) {
	e.ControllerName = c.Name
	e.DoorName = doorNameMap[e.DoorNumber]
	e.MozuID = c.MozuID

	staff := cardStaffMap[e.CardNumber]
	if e.Company = staff.Company; len(e.Company) == 0 {
		e.Company = consts.UnknownName
	}
	if e.Username = staff.Name; len(e.Username) == 0 {
		e.Username = consts.UnknownName
	}
}
