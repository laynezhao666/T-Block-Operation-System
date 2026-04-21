// Package http 实现HTTP门禁控制器协议的驱动层。
package http

import (
	"fmt"
	time2 "time"

	"dac/entity/config"
	"dac/entity/model/driver"
	"dac/entity/utils/dhttp"
	"dac/logic/cache"

	"dac/entity/utils/ttime"
)

// Alarm HTTP协议的告警数据结构
type Alarm struct {
	Index       int    `json:"index"`    // 告警索引
	Time        string `json:"time"`     // 告警时间
	DoorNumber  int    `json:"door"`     // 门编号
	Type        int    `json:"type"`     // 告警类型
	State       int    `json:"state"`    // 告警状态
	Description string `json:"alm_desc"` // 告警描述
}

// AlarmData 告警分页数据（增量拉取模式）
type AlarmData struct {
	Offset int     `json:"alarm_last_get"` // 当前偏移量
	Last   int     `json:"alarm_final"`    // 最后索引
	Alarms []Alarm `json:"alarm"`          // 告警列表
}

// HistoryAlarmData 历史告警分页数据
type HistoryAlarmData struct {
	Offset int     `json:"alarm_last_get"` // 当前偏移量
	Last   int     `json:"alarm_count"`    // 告警总数
	Alarms []Alarm `json:"alarm"`          // 告警列表
}

// CurrentAlarmEvent 当前活跃告警事件
type CurrentAlarmEvent struct {
	Type int    `json:"type"` // 告警类型
	Desc string `json:"desc"` // 告警描述
}

// CurrentAlarmData 当前活跃告警数据（按门分组）
type CurrentAlarmData struct {
	Door   int                 `json:"door"`  // 门编号
	Alarms []CurrentAlarmEvent `json:"alarm"` // 告警事件列表
}

// getDriverAlarm 将HTTP告警转换为驱动层告警模型
func getDriverAlarm(a *Alarm) (*driver.Alarm, error) {
	t, err := ttime.ParseLocal(a.Time)
	if err != nil {
		return nil, err
	}

	return &driver.Alarm{
		Index:       a.Index,
		Timestamp:   t.Unix(),
		DoorNumber:  driver.DoorNumberType(a.DoorNumber),
		Type:        driver.AlarmType(a.Type),
		State:       driver.AlarmStateType(a.State),
		Description: a.Description,
	}, nil
}

// getAlarms 从指定URL获取告警数据并转换为驱动层模型
func (c *Controller) getAlarms(url string) (driver.AlarmData, error) {
	var data AlarmData
	err := dhttp.GetJSON(url, c.timeout, &data)
	if err != nil {
		return driver.AlarmData{}, err
	}

	alarms := make([]driver.Alarm, 0, len(data.Alarms))
	for i := range data.Alarms {
		a, err := getDriverAlarm(&data.Alarms[i])
		if err != nil {
			c.logger.Warnf("convert alarm %+v error: %v", data.Alarms[i], err)
			continue
		}

		alarms = append(alarms, *a)
	}

	return driver.AlarmData{
		Offset: data.Offset,
		Last:   data.Last,
		Alarms: alarms,
	}, nil
}

// getMDCAlarms MDC版本的告警获取（先获取总数再获取数据）
func (c *Controller) getMDCAlarms(offset interface{}) (driver.AlarmData, error) {
	var (
		url      string
		data     driver.AlarmData
		tempData driver.AlarmData
		err      error
	)

	// 第一个请求为获取总数
	url = c.urlProducer.GetMDCAlarmURL(0)

	if tempData, err = c.getAlarms(url); err != nil {
		return tempData, err
	}
	data.Last = tempData.Offset

	if o, ok := offset.(int); ok && o == 0 {
		data.Offset = 1
		return data, nil
	}

	url = c.urlProducer.GetMDCAlarmURL(offset)
	if tempData, err = c.getAlarms(url); err != nil {
		return tempData, err
	}
	data.Offset = tempData.Offset

	data.Alarms = tempData.Alarms

	return data, nil
}

// getAlarmsWithOffset 根据偏移量获取告警（自动选择MDC或标准模式）
func (c *Controller) getAlarmsWithOffset(offset interface{}) (driver.AlarmData, error) {
	info, ok := cache.Get().GetController(c.baseInfo.ID)
	if ok && config.C.UseGetEvents(info.MozuID) {
		return c.getMDCAlarms(offset)
	}

	if c.isVersionMDC {
		return c.getMDCAlarms(offset)
	}

	url := c.urlProducer.GetHistoryAlarmURL(offset)

	return c.getHistoryAlarms(url)
}

// GetAlarms 根据偏移量获取告警数据
func (c *Controller) GetAlarms(offset int) (driver.AlarmData, error) {
	return c.getAlarmsWithOffset(offset)
}

// GetAlarmsByTime 按时间区间获取历史告警（支持分页遍历）
func (c *Controller) GetAlarmsByTime(timeInterval driver.TimeInterval) (driver.AlarmData, error) {
	beginTime := time2.Unix(timeInterval.BeginTimestamp, 0)
	endTime := getFetchTime(time2.Unix(timeInterval.EndTimestamp, 0))
	if !beginTime.Before(endTime) {
		return driver.AlarmData{}, fmt.Errorf("invalid time interval: %+v, begin time: %v, end time: %v",
			timeInterval, beginTime, endTime)
	}

	beg := FormatTime(beginTime)
	end := FormatTime(endTime)

	var result driver.AlarmData
	result.EndTimestamp = endTime.Unix()

	url := c.urlProducer.GetHistoryAlarmByTimestampURL(beg, end, 0)
	data, err := c.getHistoryAlarms(url)
	if err != nil {
		return result, err
	}
	result.Alarms = make([]driver.Alarm, 0, len(data.Alarms))
	// 此处 alarm 中的 index 无实际意义
	result.Alarms = append(result.Alarms, data.Alarms...)

	last := data.Last
	for index := data.Offset; index < last; {
		url = c.urlProducer.GetHistoryAlarmByTimestampURL(beg, end, index)
		if data, err = c.getHistoryAlarms(url); err != nil {
			return result, err
		}
		result.Alarms = append(result.Alarms, data.Alarms...)
		index = data.Offset

		time2.Sleep(c.alarmFetchWaitTime)
	}

	return result, nil
}

// GetAlarmsWhenVerify 校验模式下获取告警数据
func (c *Controller) GetAlarmsWhenVerify(offset interface{}) (driver.AlarmData, error) {
	return c.getAlarmsWithOffset(offset)
}

// getHistoryAlarms 从指定URL获取历史告警数据
func (c *Controller) getHistoryAlarms(url string) (driver.AlarmData, error) {
	var data HistoryAlarmData
	err := dhttp.GetJSON(url, c.timeout, &data)
	if err != nil {
		return driver.AlarmData{}, err
	}

	alarms := make([]driver.Alarm, 0, len(data.Alarms))
	for i := range data.Alarms {
		// 该接口返回数据中无 index 字段，手动计算
		data.Alarms[i].Index = data.Offset - i
		a, err := getDriverAlarm(&data.Alarms[i])
		if err != nil {
			c.logger.Warnf("convert alarm %+v error: %v", data.Alarms[i], err)
			continue
		}

		alarms = append(alarms, *a)
	}

	return driver.AlarmData{
		Offset: data.Offset,
		Last:   data.Last,
		Alarms: alarms,
	}, nil
}

// GetCurrentAlarm 获取当前活跃的告警列表
func (c *Controller) GetCurrentAlarm() ([]driver.CurrentAlarmData, error) {
	ignoreError := c.isVersionMDC || c.isVersion1
	var data []CurrentAlarmData
	url := c.urlProducer.GetCurrentAlarmURL()
	err := dhttp.GetJSON(url, c.timeout, &data)
	if err != nil {
		if !ignoreError {
			return nil, err
		}
		c.filterLogger.Infof("GetCurrentAlarm", "ignore get current alarm error: %v", err)
	}

	results := make([]driver.CurrentAlarmData, 0, len(data))
	for i := range data {
		alarms := make([]driver.CurrentAlarmEvent, 0, len(data[i].Alarms))
		for j := range data[i].Alarms {
			a := &data[i].Alarms[j]
			alarms = append(alarms, driver.CurrentAlarmEvent{
				Type: a.Type,
				Desc: a.Desc,
			})
		}

		results = append(results, driver.CurrentAlarmData{
			Door:   data[i].Door,
			Alarms: alarms,
		})
	}

	return results, nil
}
