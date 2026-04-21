// Package controller 提供门禁控制器的日志工具方法。
package controller

// Infof 输出Info级别日志
func (c *Controller) Infof(format string, args ...interface{}) {
	c.logger.Infof(format, args...)
}

// Warnf 输出Warn级别日志
func (c *Controller) Warnf(format string, args ...interface{}) {
	c.logger.Warnf(format, args...)
}

// Errorf 输出Error级别日志
func (c *Controller) Errorf(format string, args ...interface{}) {
	c.logger.Errorf(format, args...)
}
