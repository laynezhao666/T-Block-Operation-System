package flog

import (
	"log"
)

var (
	d *defaultLogger = &defaultLogger{}
)

type defaultLogger struct {
	Logger
}

// Debug ...
func (l *defaultLogger) Debug(args ...interface{}) {
	log.Default().Print(args...)
}

// Debugf flog
func (l *defaultLogger) Debugf(format string, args ...interface{}) {
	log.Default().Printf(format, args...)
}

// Info ...
func (l *defaultLogger) Info(args ...interface{}) {
	log.Default().Print(args...)
}

// Infof flog
func (l *defaultLogger) Infof(format string, args ...interface{}) {
	log.Default().Printf(format, args...)
}

// Warn ...
func (l *defaultLogger) Warn(args ...interface{}) {
	log.Default().Print(args...)
}

// Warnf flog
func (l *defaultLogger) Warnf(format string, args ...interface{}) {
	log.Default().Printf(format, args...)
}

// Error flog
func (l *defaultLogger) Error(args ...interface{}) {
	log.Default().Print(args...)
}

// Errorf flog
func (l *defaultLogger) Errorf(format string, args ...interface{}) {
	log.Default().Printf(format, args...)
}
