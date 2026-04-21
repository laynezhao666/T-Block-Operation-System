// Package rrpc 提供请求-响应式RPC管理器，支持异步等待响应。
package rrpc

import (
	"sync"
	"time"
)

// globalManager 全局结果管理器单例
var (
	globalManager *resultManager
)

// channelElem 通道元素类型
type channelElem struct{}

// channelType 通知通道类型
type channelType chan channelElem

// channelMap 消息ID到通知通道的映射
type channelMap map[string]channelType

// valueType 存储响应值的包装类型
type valueType struct {
	value interface{}
}

// valueMap 消息ID到响应值的映射
type valueMap map[string]*valueType

// resultManager 管理RRPC请求的异步结果等待和通知
type resultManager struct {
	mutex    sync.RWMutex
	channels channelMap
	values   valueMap
}

// init 初始化全局结果管理器并启动后台清理协程
func init() {
	globalManager = &resultManager{
		channels: make(channelMap),
		values:   make(valueMap),
	}
	go globalManager.loop()
}

// Manager 获取全局结果管理器实例
func Manager() *resultManager {
	return globalManager
}

// Get 尝试在 timeout 时间内获取 key 对应的值。
func (m *resultManager) Get(
	key string, timeout time.Duration,
) (interface{}, bool) {
	if m == nil {
		return nil, false
	}

	ch, ok := m.createChannel(key)
	if !ok {
		return nil, false
	}

	defer m.pruneKey(key)

	select {
	case <-ch:
		return m.getValue(key)
	case <-time.After(timeout):
		return nil, false
	}
}

// Set 设置 key 对应的值，并通知等待方
func (m *resultManager) Set(key string, value interface{}) {
	if m == nil {
		return
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	ch, ok := m.channels[key]
	if !ok {
		return
	}

	m.values[key] = &valueType{value: value}
	ch <- channelElem{}
}

// getValue 获取指定key的响应值
func (m *resultManager) getValue(
	key string,
) (interface{}, bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	v, ok := m.values[key]
	if v == nil {
		return nil, false
	}
	return v.value, ok
}

// getChannel 获取指定key的通知通道
func (m *resultManager) getChannel(
	key string,
) (channelType, bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	ch, ok := m.channels[key]
	return ch, ok
}

// createChannel 为指定key创建通知通道，若已存在则返回false
func (m *resultManager) createChannel(
	key string,
) (channelType, bool) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	_, ok := m.channels[key]
	if ok {
		return nil, false
	}

	ch := make(channelType, 1)
	m.channels[key] = ch

	return ch, true
}

// pruneKey 清理指定key的通道和值
func (m *resultManager) pruneKey(key string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	ch, ok := m.channels[key]
	if ok {
		close(ch)
		delete(m.channels, key)
	}
	delete(m.values, key)
}

// loop 后台定时清理协程，每小时执行一次内存收缩
func (m *resultManager) loop() {
	for {
		select {
		case <-time.After(time.Hour):
			m.shrink()
		}
	}
}

// shrink 收缩内部map，释放多余的内存
func (m *resultManager) shrink() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	tempChannels := make(channelMap, len(m.channels))
	for k, ch := range m.channels {
		tempChannels[k] = ch
	}
	m.channels = tempChannels

	tempValues := make(valueMap, len(m.values))
	for k, v := range m.values {
		tempValues[k] = v
	}
	m.values = tempValues
}
