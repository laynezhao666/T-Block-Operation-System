// Package xbrother 实现XBrother门禁控制器协议的驱动层。
package xbrother

import (
	"dac/entity/model/driver/xbrother"
)

// Reset 重置控制器（关闭火警报警）
func (c *Controller) Reset() error {
	_, err := c.setFireAlarm(xbrother.AlarmSettingReq{
		DisableAlarm:    1,
		KeepEnableAlarm: 0,
	}, 0)
	return err
}
