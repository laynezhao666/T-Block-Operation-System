// Package cacs 实现 CACS 协议门控器驱动。
package cacs

import (
	"context"
	"fmt"

	"dac/entity/model/driver"
	"dac/entity/model/driver/cacs"
	"dac/logic/collect/driver/cacs/consts"
	"dac/repo/dac"
)

// getAlarmWithOffset 根据偏移量从数据库分页读取告警数据。
func (c *Controller) getAlarmWithOffset(
	offsetInterface interface{},
) (driver.AlarmData, error) {
	var res driver.AlarmData
	offset, ok := offsetInterface.(int)
	if !ok {
		return res, fmt.Errorf("offset type error")
	}

	// 从数据库读取告警
	totalCount, driverAlarms, err := dac.GetRW().GetDriverAlarms(
		context.Background(), c.baseInfo.ID,
		c.chanInfo.ChannelID, offset, consts.DefaultLimit)
	if err != nil {
		return res, err
	}

	res = driver.AlarmData{
		Offset: offset + len(driverAlarms),
		Last:   int(totalCount),
		Alarms: make([]driver.Alarm, len(driverAlarms)),
	}

	for i := range driverAlarms {
		alarm := &driverAlarms[i]
		res.Alarms[i] = driver.Alarm{
			Index:       alarm.Index,
			Timestamp:   alarm.Timestamp,
			DoorNumber:  driver.DoorNumberType(alarm.DoorNumber),
			Type:        driver.AlarmType(alarm.Type),
			State:       driver.AlarmStateType(alarm.State),
			Description: alarm.Description,
		}
	}
	return res, nil
}

// GetAlarms 获取门控器告警数据，从指定偏移量开始分页读取。
func (c *Controller) GetAlarms(offset int) (driver.AlarmData, error) {
	if _, err := c.checkConnection(); err != nil {
		return driver.AlarmData{}, err
	}
	return c.getAlarmWithOffset(offset)
}

// GetAlarmsByTime 按时间区间获取告警数据（CACS协议暂未实现）。
func (c *Controller) GetAlarmsByTime(
	_ driver.TimeInterval,
) (driver.AlarmData, error) {
	// TODO: implement GetAlarmsByTime for CACS protocol
	return driver.AlarmData{}, nil
}

// GetAlarmsWhenVerify 校验时获取告警数据，用于数据一致性校验。
func (c *Controller) GetAlarmsWhenVerify(
	offset interface{},
) (driver.AlarmData, error) {
	if _, err := c.checkConnection(); err != nil {
		return driver.AlarmData{}, err
	}
	return c.getAlarmWithOffset(offset)
}

// GetCurrentAlarm 获取当前实时告警数据。
// 从 DoorServer 的 currentAlarms 缓存中读取各门的告警状态。
func (c *Controller) GetCurrentAlarm() ([]driver.CurrentAlarmData, error) {
	server, err := c.checkConnection()
	if err != nil {
		return nil, err
	}
	res := make([]driver.CurrentAlarmData, 0)
	var m map[uint32]map[uint8]cacs.EventAlarmItem
	func() {
		server.uploadEventAlarmMutex.RLock()
		defer server.uploadEventAlarmMutex.RUnlock()
		m = server.currentAlarms
	}()
	for doorId, alarms := range m {
		var currentAlarmData driver.CurrentAlarmData
		currentAlarmData.Door = int(doorId)
		currentAlarmData.Alarms = make([]driver.CurrentAlarmEvent, 0)
		for alarmType := range alarms {
			currentAlarmData.Alarms = append(
				currentAlarmData.Alarms,
				driver.CurrentAlarmEvent{
					Type: int(alarmType),
					Desc: consts.EventAlarmInfoMap[alarmType],
				})
		}
		res = append(res, currentAlarmData)
	}
	return res, nil
}
