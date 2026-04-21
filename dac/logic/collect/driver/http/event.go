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

// Event HTTP协议的事件数据结构
type Event struct {
	Index       int    `json:"index"`     // 事件索引
	Time        string `json:"time"`      // 事件时间
	CardNumber  string `json:"card_no"`   // 卡号
	UserName    string `json:"user_name"` // 用户名
	DoorNumber  int    `json:"door"`      // 门编号
	Direction   int    `json:"direction"` // 进出方向
	Type        int    `json:"type"`      // 事件类型
	Description string `json:"desc"`      // 事件描述
}

// EventData 事件分页数据（增量拉取模式）
type EventData struct {
	Offset int     `json:"record_last_get"` // 当前偏移量
	Last   int     `json:"record_final"`    // 最后记录索引
	Events []Event `json:"record"`          // 事件列表
}

// HistoryEventData 历史事件分页数据
type HistoryEventData struct {
	Offset int     `json:"record_last_get"` // 当前偏移量
	Last   int     `json:"record_count"`    // 记录总数
	Events []Event `json:"record"`          // 事件列表
}

// getDriverEvent 将HTTP事件转换为驱动层事件模型
func getDriverEvent(e *Event) (*driver.Event, error) {
	t, err := ttime.ParseLocal(e.Time)
	if err != nil {
		return nil, err
	}

	return &driver.Event{
		Index:       e.Index,
		Timestamp:   t.Unix(),
		CardNumber:  e.CardNumber,
		Username:    e.UserName,
		DoorNumber:  driver.DoorNumberType(e.DoorNumber),
		Direction:   driver.DirectionType(e.Direction),
		Type:        driver.EventType(e.Type),
		Description: e.Description,
	}, nil
}

// getEvents 从指定URL获取事件数据并转换为驱动层模型
func (c *Controller) getEvents(url string) (driver.EventData, error) {
	var data EventData
	err := dhttp.GetJSON(url, c.timeout, &data)
	if err != nil {
		return driver.EventData{}, err
	}

	events := make([]driver.Event, 0, len(data.Events))
	for i := range data.Events {
		e, err := getDriverEvent(&data.Events[i])
		if err != nil {
			c.logger.Warnf("convert event %+v error: %v", data.Events[i], err)
			continue
		}

		events = append(events, *e)
	}

	return driver.EventData{
		Offset: data.Offset,
		Last:   data.Last,
		Events: events,
	}, nil
}

// getMDCEvents MDC版本的事件获取（先获取总数再获取数据）
func (c *Controller) getMDCEvents(offset interface{}) (driver.EventData, error) {
	var (
		url      string
		data     driver.EventData
		tempData driver.EventData
		err      error
	)

	// 第一个请求为获取总数
	url = c.urlProducer.GetMDCEventURL(0)

	if tempData, err = c.getEvents(url); err != nil {
		return tempData, err
	}
	data.Last = tempData.Offset

	if o, ok := offset.(int); ok && o == 0 {
		data.Offset = 1
		return data, nil
	}

	url = c.urlProducer.GetMDCEventURL(offset)
	if tempData, err = c.getEvents(url); err != nil {
		return tempData, err
	}
	data.Offset = tempData.Offset

	data.Events = tempData.Events

	return data, nil
}

// 检查缓存、是否MDC版本和 URL 来决定从哪里获取事件数据
func (c *Controller) getEventsWithOffset(offset interface{}) (driver.EventData, error) {
	info, ok := cache.Get().GetController(c.baseInfo.ID)
	if ok && config.C.UseGetEvents(info.MozuID) {
		return c.getMDCEvents(offset)
	}

	if c.isVersionMDC {
		return c.getMDCEvents(offset)
	}

	url := c.urlProducer.GetHistoryEventHisURL(offset)
	return c.getHistoryEvents(url)
}

// GetEvents 根据偏移量获取事件数据
func (c *Controller) GetEvents(offset int) (driver.EventData, error) {
	return c.getEventsWithOffset(offset)
}

// GetEventsByTime 按时间区间获取历史事件（支持分页遍历）
func (c *Controller) GetEventsByTime(timeInterval driver.TimeInterval) (driver.EventData, error) {
	beginTime := time2.Unix(timeInterval.BeginTimestamp, 0)
	endTime := getFetchTime(time2.Unix(timeInterval.EndTimestamp, 0))
	if !beginTime.Before(endTime) {
		return driver.EventData{}, fmt.Errorf("invalid time interval: %+v, begin time: %v, end time: %v",
			timeInterval, beginTime, endTime)
	}

	beg := FormatTime(beginTime)
	end := FormatTime(endTime)

	var result driver.EventData
	result.EndTimestamp = endTime.Unix()

	url := c.urlProducer.GetHistoryEventByTimestampURL(beg, end, 0)
	data, err := c.getHistoryEvents(url)
	if err != nil {
		return result, err
	}
	result.Events = make([]driver.Event, 0, len(data.Events))
	// 此处 event 中的 index 无实际意义
	result.Events = append(result.Events, data.Events...)

	last := data.Last
	for index := data.Offset; index < last; {
		url = c.urlProducer.GetHistoryEventByTimestampURL(beg, end, index)
		if data, err = c.getHistoryEvents(url); err != nil {
			return result, err
		}
		result.Events = append(result.Events, data.Events...)
		index = data.Offset

		time2.Sleep(c.eventFetchWaitTime)
	}

	return result, nil
}

// GetEventsWhenVerify 校验模式下获取事件数据
func (c *Controller) GetEventsWhenVerify(offset interface{}) (driver.EventData, error) {
	return c.getEventsWithOffset(offset)
}

// getHistoryEvents 从指定URL获取历史事件数据
func (c *Controller) getHistoryEvents(url string) (driver.EventData, error) {
	var data HistoryEventData
	err := dhttp.GetJSON(url, c.timeout, &data)
	if err != nil {
		return driver.EventData{}, err
	}

	events := make([]driver.Event, 0, len(data.Events))
	for i := range data.Events {
		// 该接口返回数据中无 index 字段，手动计算
		data.Events[i].Index = data.Offset - i
		e, err := getDriverEvent(&data.Events[i])
		if err != nil {
			c.logger.Warnf("convert event %+v error: %v", data.Events[i], err)
			continue
		}

		events = append(events, *e)
	}

	return driver.EventData{
		Offset: data.Offset,
		Last:   data.Last,
		Events: events,
	}, nil
}
