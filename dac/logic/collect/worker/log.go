// Package worker 提供门禁控制器的工作通道管理。
package worker

// Infof 输出带通道前缀的Info级别日志
func (c *Channel) Infof(format string, args ...interface{}) {
	c.logger.Infof(format, args...)
}

// Warnf 输出带通道前缀的Warn级别日志
func (c *Channel) Warnf(format string, args ...interface{}) {
	c.logger.Warnf(format, args...)
}
