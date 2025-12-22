package model

import (
	"sync"
)

// AvailableChannelIndex 可用通道索引
type AvailableChannelIndex struct {
	// 从前到后，首个可用索引，索引从 0 开始
	firstIndex int
	mutex      sync.RWMutex
}

// NewAvailableChannelIndex NewAvailableChannelIndex
func NewAvailableChannelIndex() *AvailableChannelIndex {
	return &AvailableChannelIndex{
		firstIndex: -1,
	}
}

// Set 设置当前首个可用索引
func (a *AvailableChannelIndex) Set(index int) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	a.firstIndex = index
}

// Get 返回当前首个可用索引，及该索引是否稳定可用（非首次可用）
func (a *AvailableChannelIndex) Get() int {
	return a.firstIndex
}
