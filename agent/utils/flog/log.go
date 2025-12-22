package flog

import (
	"fmt"
	"sync"
	"time"
)

type Level uint8

const (
	levelDebug Level = iota
	levelInfo
	levelWarn
	levelError
)

// Logger 日志接口
type Logger interface {
	Debug(args ...interface{})
	Debugf(format string, args ...interface{})
	Info(args ...interface{})
	Infof(format string, args ...interface{})
	Warn(args ...interface{})
	Warnf(format string, args ...interface{})
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
}

// Filter 日志过滤器
type Filter struct {
	sync.RWMutex
	interval   time.Duration
	logger     Logger
	entry      map[interface{}]struct{}
	stopCalled bool
	stopChan   chan struct{}
	prefix     string
}

// NewDefaultFilterLogger 默认日志过滤器
func NewDefaultFilterLogger(interval time.Duration) *Filter {
	return NewFilterLogger(interval, d)
}

// NewFilterLogger 日志过滤器
func NewFilterLogger(interval time.Duration, logger Logger) *Filter {
	f := &Filter{
		interval:   interval,
		entry:      make(map[interface{}]struct{}),
		stopCalled: false,
		stopChan:   make(chan struct{}, 1),
		logger:     logger,
		prefix:     fmt.Sprintf("[every %v] ", interval),
	}
	go f.refresh()
	return f
}

// Insert 插入日志过滤器
func (f *Filter) Insert(key interface{}) bool {
	if f == nil {
		return false
	}

	f.Lock()
	defer f.Unlock()
	if f.contain(key) {
		return false
	}

	f.insert(key)
	return true
}

// Debug 调试日志
func (f *Filter) Debug(key interface{}, args ...interface{}) {
	f.log(levelDebug, key, args...)
}

// Debugf 调试日志
func (f *Filter) Debugf(key interface{}, format string, args ...interface{}) {
	f.logf(levelDebug, key, format, args...)
}

// Info 信息日志
func (f *Filter) Info(key interface{}, args ...interface{}) {
	f.log(levelInfo, key, args...)
}

// Infof 信息日志
func (f *Filter) Infof(key interface{}, format string, args ...interface{}) {
	f.logf(levelInfo, key, format, args...)
}

// Warn 警告日志
func (f *Filter) Warn(key interface{}, args ...interface{}) {
	f.log(levelWarn, key, args...)
}

// Warnf 警告日志
func (f *Filter) Warnf(key interface{}, format string, args ...interface{}) {
	f.logf(levelWarn, key, format, args...)
}

// Error 错误日志
func (f *Filter) Error(key interface{}, args ...interface{}) {
	f.log(levelError, key, args...)
}

// Errorf 错误日志
func (f *Filter) Errorf(key interface{}, format string, args ...interface{}) {
	f.logf(levelError, key, format, args...)
}

func (f *Filter) logf(level Level, key interface{}, format string, v ...interface{}) {
	if f == nil {
		return
	}

	f.RLock()
	if f.contain(key) {
		f.RUnlock()
		return
	}
	f.RUnlock()

	f.Lock()
	if f.contain(key) {
		f.Unlock()
		return
	}
	f.insert(key)
	f.Unlock()

	switch level {
	case levelDebug:
		f.logger.Debugf(f.prefix+format, v...)
	case levelInfo:
		f.logger.Infof(f.prefix+format, v...)
	case levelWarn:
		f.logger.Warnf(f.prefix+format, v...)
	case levelError:
		f.logger.Errorf(f.prefix+format, v...)
	default:
		break
	}
}

func (f *Filter) log(level Level, key interface{}, v ...interface{}) {
	if f == nil {
		return
	}

	f.RLock()
	if f.contain(key) {
		f.RUnlock()
		return
	}
	f.RUnlock()

	f.Lock()
	if f.contain(key) {
		f.Unlock()
		return
	}
	f.insert(key)
	f.Unlock()

	newArgs := make([]interface{}, 0, len(v)+1)
	newArgs = append(newArgs, f.prefix)
	newArgs = append(newArgs, v...)

	switch level {
	case levelDebug:
		f.logger.Debug(newArgs...)
	case levelInfo:
		f.logger.Info(newArgs...)
	case levelWarn:
		f.logger.Warn(newArgs...)
	case levelError:
		f.logger.Error(newArgs...)
	default:
		break
	}
}

// Stop 停止过滤器
func (f *Filter) Stop() {
	if f == nil {
		return
	}

	f.Lock()
	defer f.Unlock()

	if f.stopCalled {
		return
	}

	f.stopCalled = true
	f.stopChan <- struct{}{}
}

func (f *Filter) contain(key interface{}) bool {
	_, ok := f.entry[key]
	return ok
}

func (f *Filter) insert(key interface{}) {
	f.entry[key] = struct{}{}
}

func (f *Filter) clear() {
	f.entry = make(map[interface{}]struct{})
}

func (f *Filter) refresh() {
	for {
		select {
		case <-time.After(f.interval):
			break
		case <-f.stopChan:
			return
		}

		f.Lock()
		f.clear()
		f.Unlock()
	}
}
