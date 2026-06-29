package iec104

import (
	"fmt"
	"time"
)

// Logger 日志引擎interface
type Logger interface {
	Error(args ...interface{})
	Errorf(format string, v ...interface{})
	Debug(args ...interface{})
	Debugf(format string, v ...interface{})
	Info(args ...interface{})
	Infof(format string, v ...interface{})
	Warn(args ...interface{})
	Warnf(format string, v ...interface{})
}

// LogFmt 默认日志引擎
type LogFmt struct {
}

// Debug Debug
func (l *LogFmt) Debug(args ...interface{}) {
	fmt.Printf("[DEBUG]time:%s, ", time.Now().Format(time.RFC3339))
	fmt.Println(args...)
}

// Debugf Debugf
func (l *LogFmt) Debugf(format string, v ...interface{}) {
	fmt.Printf("[DEBUG]time:%s, ", time.Now().Format(time.RFC3339))
	fmt.Printf(format+"\n", v...)
}

// Info Info
func (l *LogFmt) Info(args ...interface{}) {
	fmt.Printf("[INFO]time:%s, ", time.Now().Format(time.RFC3339))
	fmt.Println(args...)
}

// Infof Infof
func (l *LogFmt) Infof(format string, v ...interface{}) {
	fmt.Printf("[INFO]time:%s, ", time.Now().Format(time.RFC3339))
	fmt.Printf(format+"\n", v...)
}

// Warn Warn
func (l *LogFmt) Warn(args ...interface{}) {
	fmt.Printf("[WARN]time:%s, ", time.Now().Format(time.RFC3339))
	fmt.Println(args...)
}

// Warnf Warnf
func (l *LogFmt) Warnf(format string, v ...interface{}) {
	fmt.Printf("[WARN]time:%s, ", time.Now().Format(time.RFC3339))
	fmt.Printf(format+"\n", v...)
}

// Error Error
func (l *LogFmt) Error(args ...interface{}) {
	fmt.Printf("[ERR]time:%s, ", time.Now().Format(time.RFC3339))
	fmt.Println(args...)
}

// Errorf Errorf
func (l *LogFmt) Errorf(format string, v ...interface{}) {
	fmt.Printf("[ERR]time:%s, ", time.Now().Format(time.RFC3339))
	fmt.Printf(format+"\n", v...)
}
