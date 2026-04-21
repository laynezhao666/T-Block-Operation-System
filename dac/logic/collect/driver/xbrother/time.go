// Package xbrother 实现XBrother门禁控制器协议的驱动层。
package xbrother

import (
	"time"

	"dac/entity/model/driver/xbrother"
	"dac/logic/collect/driver/xbrother/consts"
)

// GetTime 获取门控器时间。
// 如果已成功同步过时间，直接返回当前时间；否则先同步再返回。
func (c *Controller) GetTime() (string, error) {
	req, now := getNowTime()
	if c.hasSetTime {
		// 如果当前启动的Controller已经成功进行了一次时间同步，默认时间已经同步，返回当前时间
		return now.String(), nil
	}
	_, err := c.setTime(req, 0)
	if err != nil {
		return "", err
	}
	return now.String(), nil
}

// SetTime 将本地时间同步到门控器
func (c *Controller) SetTime() error {
	req, _ := getNowTime()
	_, err := c.setTime(req, 0)
	if err != nil {
		return err
	}
	c.hasSetTime = true
	return nil
}

// setTime 发送设置时间请求到门控器
func (c *Controller) setTime(
	req xbrother.SetTimeReq, doorNo uint8,
) (xbrother.CommonResp, error) {
	return c.sendRequest(
		req, doorNo,
		consts.GetRRPCSetTime(c.chanInfo.ChannelID),
		consts.CommandSetTime)
}

// getNowTime 获取当前时间并构造SetTimeReq请求
func getNowTime() (xbrother.SetTimeReq, time.Time) {
	now := time.Now()
	req := xbrother.SetTimeReq{
		Second: uint8(now.Second()),
		Minute: uint8(now.Minute()),
		Hour:   uint8(now.Hour()),
		Week:   uint8(now.Weekday() + 1), // 协议规定7代表星期六
		Day:    uint8(now.Day()),
		Month:  uint8(now.Month()),
		Year:   uint8(now.Year() - 2000), // 协议规定真实年份等于Year+2000
	}
	return req, now
}
