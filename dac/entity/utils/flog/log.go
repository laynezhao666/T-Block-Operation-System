// Package flog 提供带频率过滤的日志工具，避免高频重复日志刷屏。
package flog

import (
	"context"
	"fmt"
	"sync"
	"time"

	"dac/entity/utils/tlog"
)

// Level 日志级别类型
type Level uint8

// 日志级别常量定义
const (
	levelDebug Level = iota
	levelInfo
	levelWarn
	levelError
)

// Filter 带频率过滤的日志器。
// 在指定时间间隔内，相同key的日志只输出一次。
type Filter struct {
	sync.RWMutex
	interval   time.Duration
	logger     tlog.Logger
	entry      map[interface{}]struct{}
	stopCalled bool
	stopChan   chan struct{}
	prefix     string
}

// NewDefaultFilterLogger 使用默认日志器创建过滤日志器
func NewDefaultFilterLogger(interval time.Duration) *Filter {
	return NewFilterLogger(interval, d)
}

// NewFilterLoggerWithContext 使用指定上下文和日志器创建过滤日志器
func NewFilterLoggerWithContext(ctx context.Context, interval time.Duration, logger tlog.Logger) *Filter {
	f := &Filter{
		interval:   interval,
		entry:      make(map[interface{}]struct{}),
		stopCalled: false,
		stopChan:   make(chan struct{}, 1),
		logger:     logger,
		prefix:     fmt.Sprintf("[every %v] ", interval),
	}
	go f.refresh(ctx)
	return f
}

// NewFilterLogger 使用指定日志器创建过滤日志器
func NewFilterLogger(interval time.Duration, logger tlog.Logger) *Filter {
	return NewFilterLoggerWithContext(context.Background(), interval, logger)
}

// Insert 插入一个key到过滤集合，若已存在返回false
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

// Debug 输出Debug级别的过滤日志
func (f *Filter) Debug(key interface{}, args ...interface{}) {
	f.log(levelDebug, key, args...)
}

// Debugf 输出格式化Debug级别的过滤日志
func (f *Filter) Debugf(key interface{}, format string, args ...interface{}) {
	f.logf(levelDebug, key, format, args...)
}

// Info 输出Info级别的过滤日志
func (f *Filter) Info(key interface{}, args ...interface{}) {
	f.log(levelInfo, key, args...)
}

// Infof 输出格式化Info级别的过滤日志
func (f *Filter) Infof(key interface{}, format string, args ...interface{}) {
	f.logf(levelInfo, key, format, args...)
}

// Warn 输出Warn级别的过滤日志
func (f *Filter) Warn(key interface{}, args ...interface{}) {
	f.log(levelWarn, key, args...)
}

// Warnf 输出格式化Warn级别的过滤日志
func (f *Filter) Warnf(key interface{}, format string, args ...interface{}) {
	f.logf(levelWarn, key, format, args...)
}

// Error 输出Error级别的过滤日志
func (f *Filter) Error(key interface{}, args ...interface{}) {
	f.log(levelError, key, args...)
}

// Errorf 输出格式化Error级别的过滤日志
func (f *Filter) Errorf(key interface{}, format string, args ...interface{}) {
	f.logf(levelError, key, format, args...)
}

// logf 内部格式化日志方法，带去重过滤
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

// log 内部日志方法，带去重过滤
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

// Stop 停止过滤日志器的后台刷新协程
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

// contain 检查key是否已在过滤集合中
func (f *Filter) contain(key interface{}) bool {
	_, ok := f.entry[key]
	return ok
}

// insert 将key插入过滤集合
func (f *Filter) insert(key interface{}) {
	f.entry[key] = struct{}{}
}

// clear 清空过滤集合
func (f *Filter) clear() {
	f.entry = make(map[interface{}]struct{})
}

// refresh 后台定时清空过滤集合，恢复日志输出能力
func (f *Filter) refresh(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
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
