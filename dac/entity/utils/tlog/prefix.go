// Package tlog 定义门禁系统通用日志接口。
package tlog

// PrefixLogger 带前缀的日志器，自动在每条日志前添加指定前缀
type PrefixLogger struct {
	prefix string
	logger Logger
}

// NewPrefixLogger 创建带前缀的日志器
func NewPrefixLogger(p string, l Logger) *PrefixLogger {
	logger := new(PrefixLogger)
	logger.prefix = p
	logger.logger = l
	return logger
}

// getArgs 在参数列表前插入前缀
func (p *PrefixLogger) getArgs(args ...interface{}) []interface{} {
	newArgs := make([]interface{}, 0, len(args)+1)
	newArgs = append(newArgs, p.prefix)
	newArgs = append(newArgs, args...)
	return newArgs
}

// getFormat 在格式字符串前添加前缀
func (p *PrefixLogger) getFormat(format string) string {
	return p.prefix + " " + format
}

// Debug 输出带前缀的Debug日志
func (p *PrefixLogger) Debug(args ...interface{}) {
	p.logger.Debug(p.getArgs(args...)...)
}

// Debugf 输出带前缀的格式化Debug日志
func (p *PrefixLogger) Debugf(format string, args ...interface{}) {
	p.logger.Debugf(p.getFormat(format), args...)
}

// Info 输出带前缀的Info日志
func (p *PrefixLogger) Info(args ...interface{}) {
	p.logger.Info(p.getArgs(args...)...)
}

// Infof 输出带前缀的格式化Info日志
func (p *PrefixLogger) Infof(format string, args ...interface{}) {
	p.logger.Infof(p.getFormat(format), args...)
}

// Warn 输出带前缀的Warn日志
func (p *PrefixLogger) Warn(args ...interface{}) {
	p.logger.Warn(p.getArgs(args...)...)
}

// Warnf 输出带前缀的格式化Warn日志
func (p *PrefixLogger) Warnf(format string, args ...interface{}) {
	p.logger.Warnf(p.getFormat(format), args...)
}

// Error 输出带前缀的Error日志
func (p *PrefixLogger) Error(args ...interface{}) {
	p.logger.Error(p.getArgs(args...)...)
}

// Errorf 输出带前缀的格式化Error日志
func (p *PrefixLogger) Errorf(format string, args ...interface{}) {
	p.logger.Errorf(p.getFormat(format), args...)
}
