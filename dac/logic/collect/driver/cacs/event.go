// Package cacs 实现CACS门禁控制器协议的驱动层。
package cacs

import (
	"context"
	"dac/entity/model/driver"
	"dac/entity/utils"
	"dac/logic/collect/driver/cacs/consts"
	"dac/repo/dac"
	"fmt"
)

// getEventsWithOffset 根据偏移量从数据库获取事件列表并转换为驱动模型
func (c *Controller) getEventsWithOffset(offsetInterface interface{}) (driver.EventData, error) {
	offset, ok := offsetInterface.(int)
	if !ok {
		return driver.EventData{}, fmt.Errorf("offset type error")
	}

	// 从数据库读取事件
	totalCount, driverEvents, err := dac.GetRW().GetDriverEvents(context.Background(), c.baseInfo.ID, c.chanInfo.ChannelID,
		offset, consts.DefaultLimit)
	if err != nil {
		return driver.EventData{}, err
	}

	return driver.EventData{
		Offset: offset + len(driverEvents),
		Last:   int(totalCount),
		Events: utils.ConvertDBEventsToDriver(driverEvents),
	}, nil
}

// GetEvents 获取事件记录（需先检查连接状态）
func (c *Controller) GetEvents(offset int) (driver.EventData, error) {
	if _, err := c.checkConnection(); err != nil {
		return driver.EventData{}, err
	}
	return c.getEventsWithOffset(offset)
}

// GetEventsByTime 按时间范围获取事件记录（暂未实现）
func (c *Controller) GetEventsByTime(_ driver.TimeInterval) (driver.EventData, error) {
	// TODO: implement GetEventsByTime for CACS protocol
	return driver.EventData{}, nil
}

// GetEventsWhenVerify 验证时获取事件（需先检查连接状态）
func (c *Controller) GetEventsWhenVerify(offset interface{}) (driver.EventData, error) {
	if _, err := c.checkConnection(); err != nil {
		return driver.EventData{}, err
	}
	return c.getEventsWithOffset(offset)
}
