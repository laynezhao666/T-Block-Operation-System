// Package cacs 实现CACS门禁控制器协议的驱动层。
package cacs

import (
	"fmt"

	"dac/entity/config"
)

// Infof 输出Info级别日志（带控制器和通道标识前缀）
func (c *Controller) Infof(format string, args ...interface{}) {
	prefix := fmt.Sprintf(
		"[controller %v, channel: %v]: ",
		c.baseInfo.ID, c.chanInfo.ChannelID)
	config.Log.Infof(prefix+format, args...)
}

// Warnf 输出Warn级别日志（带控制器和通道标识前缀）
func (c *Controller) Warnf(format string, args ...interface{}) {
	prefix := fmt.Sprintf(
		"[controller %v, channel: %v]: ",
		c.baseInfo.ID, c.chanInfo.ChannelID)
	config.Log.Warnf(prefix+format, args...)
}

// Errorf 输出Error级别日志（带控制器和通道标识前缀）
func (c *Controller) Errorf(format string, args ...interface{}) {
	prefix := fmt.Sprintf(
		"[controller %v, channel: %v]: ",
		c.baseInfo.ID, c.chanInfo.ChannelID)
	config.Log.Errorf(prefix+format, args...)
}
