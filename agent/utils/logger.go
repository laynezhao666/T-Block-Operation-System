package utils

import (
	"trpc.group/trpc-go/trpc-go/log"
)

// Logger 日志接口
type Logger struct {
	println func(v ...interface{})
	print   func(v ...interface{})
	printf  func(format string, v ...interface{})
}

// NewLogger 创建日志接口
func NewLogger(println func(v ...interface{}), printf func(string, ...interface{})) Logger {
	return Logger{
		println: println,
		printf:  printf,
		print:   println,
	}
}

// NewDefaultDebugLogger 创建默认debug日志接口
func NewDefaultDebugLogger() Logger {
	return NewLogger(log.GetDefaultLogger().Debug, log.GetDefaultLogger().Debugf)
}

// NewDefaultInfoLogger 创建默认info日志接口
func NewDefaultInfoLogger() Logger {
	return NewLogger(log.GetDefaultLogger().Info, log.GetDefaultLogger().Infof)
}

// NewDefaultWarnLogger 创建默认warn日志接口
func NewDefaultWarnLogger() Logger {
	return NewLogger(log.GetDefaultLogger().Warn, log.GetDefaultLogger().Warnf)
}

// NewDefaultErrorLogger 创建默认error日志接口
func NewDefaultErrorLogger() Logger {
	return NewLogger(log.GetDefaultLogger().Error, log.GetDefaultLogger().Errorf)
}

// Println 打印日志
func (l Logger) Println(v ...interface{}) {
	l.println(v...)
}

// Print 打印日志
func (l Logger) Print(v ...interface{}) {
	l.print(v...)
}

// Printf 打印日志
func (l Logger) Printf(format string, v ...interface{}) {
	l.printf(format, v...)
}
