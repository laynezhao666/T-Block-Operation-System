// Package flog 提供默认日志实现，基于标准库log包。
// 当未配置外部日志框架时作为兜底日志输出。
package flog

import (
	"log"

	"dac/entity/utils/tlog"
)

// d 默认日志实例
var (
	d = &defaultLogger{}
)

// defaultLogger 基于标准库log的默认日志实现
type defaultLogger struct {
	tlog.Logger
}

// Debug 输出Debug级别日志
func (l *defaultLogger) Debug(args ...interface{}) {
	log.Default().Print(args...)
}

// Debugf 输出格式化Debug级别日志
func (l *defaultLogger) Debugf(format string, args ...interface{}) {
	log.Default().Printf(format, args...)
}

// Info 输出Info级别日志
func (l *defaultLogger) Info(args ...interface{}) {
	log.Default().Print(args...)
}

// Infof 输出格式化Info级别日志
func (l *defaultLogger) Infof(format string, args ...interface{}) {
	log.Default().Printf(format, args...)
}

// Warn 输出Warn级别日志
func (l *defaultLogger) Warn(args ...interface{}) {
	log.Default().Print(args...)
}

// Warnf 输出格式化Warn级别日志
func (l *defaultLogger) Warnf(format string, args ...interface{}) {
	log.Default().Printf(format, args...)
}

// Error 输出Error级别日志
func (l *defaultLogger) Error(args ...interface{}) {
	log.Default().Print(args...)
}

// Errorf 输出格式化Error级别日志
func (l *defaultLogger) Errorf(format string, args ...interface{}) {
	log.Default().Printf(format, args...)
}
