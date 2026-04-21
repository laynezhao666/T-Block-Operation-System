// Package utils 提供门禁系统通用工具函数。
package utils

import (
	"dac/entity/consts"
	"dac/entity/model/db"
	"dac/entity/model/driver"
	"dac/entity/model/rt"
	"dac/entity/utils/ttime"
)

// BuildDoorPoints 根据门状态和告警数据构建门监测点，供各驱动的 GetDoorPoints 复用。
// isOpenAlarm 用于判断某条告警是否属于"门开超时"告警，不同协议的告警类型值不同。
// 若 isOpenAlarm 为 nil，则认为 alarmData 中的每条记录都是有效告警（适用于 xbrother/chd806d4）。
func BuildDoorPoints(
	controllerID db.IDType,
	doors []int,
	states map[int]*rt.Point,
	alarmData []driver.CurrentAlarmData,
	isOpenAlarm func(alarmType int) bool,
) map[string]map[int]*rt.Point {
	timestamp := ttime.GetNowUTC().UnixMilli()
	alarmPoints := make(map[int]*rt.Point, len(doors))

	for i := range alarmData {
		d := alarmData[i].Door
		if isOpenAlarm == nil {
			// 无过滤函数，直接视为有告警
			p := new(rt.Point)
			p.ID = GenerateDoorOpenAlarmID(controllerID, d)
			p.SetValueWithTime(1, timestamp)
			alarmPoints[d] = p
		} else {
			for j := range alarmData[i].Alarms {
				if isOpenAlarm(alarmData[i].Alarms[j].Type) {
					p := new(rt.Point)
					p.ID = GenerateDoorOpenAlarmID(controllerID, d)
					p.SetValueWithTime(1, timestamp)
					alarmPoints[d] = p
					break
				}
			}
		}
	}

	for _, d := range doors {
		if _, ok := alarmPoints[d]; ok {
			continue
		}
		p := new(rt.Point)
		p.ID = GenerateDoorOpenAlarmID(controllerID, d)
		p.SetValueWithTime(0, timestamp)
		alarmPoints[d] = p
	}

	return map[string]map[int]*rt.Point{
		consts.StandardIDDoorState: states,
		consts.StandardIDOpenAlarm: alarmPoints,
	}
}
