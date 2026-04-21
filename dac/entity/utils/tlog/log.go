// Package tlog 定义门禁系统通用日志接口。
package tlog

// Logger 通用日志接口，支持Debug/Info/Warn/Error四个级别
type Logger interface {
	// Debug 输出Debug级别日志
	Debug(args ...interface{})
	// Debugf 输出格式化Debug级别日志
	Debugf(format string, args ...interface{})
	// Info 输出Info级别日志
	Info(args ...interface{})
	// Infof 输出格式化Info级别日志
	Infof(format string, args ...interface{})
	// Warn 输出Warn级别日志
	Warn(args ...interface{})
	// Warnf 输出格式化Warn级别日志
	Warnf(format string, args ...interface{})
	// Error 输出Error级别日志
	Error(args ...interface{})
	// Errorf 输出格式化Error级别日志
	Errorf(format string, args ...interface{})
}
