// Package http 实现HTTP门禁控制器协议的驱动层。
package http

import (
	"dac/entity/model/driver"
	"dac/entity/utils/dhttp"
)

// GetTimeGroup 从控制器获取指定编号的时间组
func (c *Controller) GetTimeGroup(timeGroup int) (driver.TimeGroup, error) {
	var data driver.TimeGroup
	url := c.urlProducer.GetTimeGroupURL(timeGroup)

	err := dhttp.GetJSON(url, c.timeout, &data)
	return data, err
}

// SetTimeGroup 设置时间组到控制器
func (c *Controller) SetTimeGroup(timeGroup driver.TimeGroup) error {
	return c.postJSON(c.urlProducer.SetTimeGroupURL(), timeGroup, nil)
}

// ClearTimeGroup 清除控制器上指定编号的时间组
func (c *Controller) ClearTimeGroup(timeGroup int) error {
	req := driver.TimeGroup{
		GroupNo: timeGroup,
	}
	return c.postJSON(c.urlProducer.ClearTimeGroupURL(), req, nil)
}
