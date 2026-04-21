// Package rtdb 提供门禁系统实时数据库的读写接口。
// 支持内存缓存和Redis两种存储后端。
package rtdb

import (
	"reflect"
	"sync"

	"dac/entity/consts"
	"dac/entity/model/rt"
)

// memoryInstance 全局内存RTDB实例
var (
	memoryInstance = Model{
		data: make(map[string]*rt.RTValue),
	}
)

// PointsUpdatedCallback 测点数据更新回调函数类型
type PointsUpdatedCallback func(rt.Points, interface{}) interface{}

// notifyObject 回调监听对象
type notifyObject struct {
	Handler PointsUpdatedCallback
	Arg     interface{}
}

// notifyObjects 回调监听对象列表
type notifyObjects []notifyObject

// Model 内存实时数据库模型，存储测点数据并支持更新通知
type Model struct {
	sync.RWMutex
	data      map[string]*rt.RTValue // 测点ID到实时值的映射
	funMutex  sync.RWMutex           // 回调列表锁
	listeners notifyObjects          // 更新回调列表
}

// NewRtdbModel 创建新的RTDB模型实例
func NewRtdbModel() *Model {
	r := new(Model)
	r.data = make(map[string]*rt.RTValue)
	return r
}

// GetAllPoints 获取所有测点数据
func (r *Model) GetAllPoints() rt.Points {
	r.RLock()
	defer r.RUnlock()

	points := make(rt.Points, 0, len(r.data))
	for id, v := range r.data {
		var p rt.Point
		p.ID = id
		p.Rtd = *v

		points = append(points, p)
	}

	return points
}

// GetPoints 批量获取指定ID的测点数据
func (r *Model) GetPoints(points rt.Points) {
	r.RLock()
	defer r.RUnlock()

	for i := range points {
		p := &points[i]

		v, ok := r.data[p.ID]
		if ok {
			p.Rtd = *v
		} else {
			p.Rtd = rt.NewRTValue()
		}
	}
}

// DeleteAllPoints 清空所有测点数据
func (r *Model) DeleteAllPoints() {
	r.Lock()
	defer r.Unlock()

	r.data = make(map[string]*rt.RTValue)
}

// SetPoints 批量设置测点数据，可选触发更新通知
func (r *Model) SetPoints(points rt.Points, notify bool) {
	r.Lock()
	for i := range points {
		p := &points[i]
		r.set(p.ID, &p.Rtd, &p.IsValueChanged)
	}
	r.Unlock()

	if notify {
		r.notify(points)
	}
}

// set 设置单个测点值，检测值是否变化
func (r *Model) set(id string, value *rt.RTValue, valChanged *bool) {
	*valChanged = false

	v, ok := r.data[id]
	if ok {
		if v.Pv != value.Pv || v.Qua != value.Qua {
			*valChanged = true
		}

		if value.Qua != consts.QualityOK {
			// 如果 qua 异常时，不会更新已存在的 pv 值。
			v.Qua = value.Qua
			v.Timestamp = value.Timestamp

			// 并修改传入的 val.pv 为已存在的 pv 值
			value.Pv = v.Pv
		} else {
			*v = *value
		}
	} else {
		r.data[id] = value
		*valChanged = true
	}
}

// UnregisterDataPointsUpdatedCallback 取消注册测点更新回调
func (r *Model) UnregisterDataPointsUpdatedCallback(handler PointsUpdatedCallback) {
	if r == nil {
		return
	}

	r.funMutex.Lock()
	defer r.funMutex.Unlock()

	rhs := reflect.ValueOf(handler).Pointer()
	index := -1
	for i, listener := range r.listeners {
		lhs := reflect.ValueOf(listener.Handler).Pointer()
		if lhs == rhs {
			index = i
			break
		}
	}

	if index >= 0 {
		newListeners := r.listeners[:index]
		r.listeners = append(newListeners, r.listeners[index+1:]...)
	}
}

// RegisterDataPointsUpdatedCallback 注册测点更新回调
func (r *Model) RegisterDataPointsUpdatedCallback(handler PointsUpdatedCallback, arg interface{}) {
	if r == nil {
		return
	}

	r.funMutex.Lock()
	defer r.funMutex.Unlock()

	rhs := reflect.ValueOf(handler).Pointer()
	for _, listener := range r.listeners {
		lhs := reflect.ValueOf(listener.Handler).Pointer()
		if lhs == rhs {
			return
		}
	}

	r.listeners = append(r.listeners, notifyObject{
		Handler: handler,
		Arg:     arg,
	})
}

// notify 触发所有已注册的更新回调
func (r *Model) notify(points rt.Points) {
	r.funMutex.RLock()
	defer r.funMutex.RUnlock()

	for _, listener := range r.listeners {
		listener.Handler(points, listener.Arg)
	}
}
