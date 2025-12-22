package worker

import (
	"sync"

	"agent/entity/definition"
)

// RequestCountMapType 请求计数
type RequestCountMapType map[definition.DeviceGidType]uint64

type requestCountMap struct {
	m RequestCountMapType
	sync.RWMutex
}

// NewRequestCountMap 创建请求计数
func NewRequestCountMap() *requestCountMap {
	return &requestCountMap{
		m: make(RequestCountMapType),
	}
}

// Increase 增加计数
func (r *requestCountMap) Increase(gid definition.DeviceGidType) {
	if r == nil {
		return
	}
	r.Lock()
	defer r.Unlock()

	r.m[gid]++
}

// Get 获取计数
func (r *requestCountMap) Get(gid definition.DeviceGidType) uint64 {
	if r == nil {
		return 0
	}
	r.RLock()
	defer r.RUnlock()

	v, ok := r.m[gid]
	if !ok {
		return 0
	}
	return v
}

// Delete 删除计数
func (r *requestCountMap) Delete(gid definition.DeviceGidType) {
	if r == nil {
		return
	}
	r.Lock()
	defer r.Unlock()

	delete(r.m, gid)
}
