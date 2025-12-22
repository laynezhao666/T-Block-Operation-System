// Package rtdb 测点实时数据库
package rtdb

import (
	"agent/utils/osal"
	"reflect"
	"sync"

	"agent/entity/consts"
	"agent/entity/definition"
	"agent/logic/collector/rtdb/model"
)

var (
	rtdbInstance rtdbModel
)

// MapData 测点数据
type MapData map[definition.DataPointIDType]*model.RTData

// DataPointsUpdatedFun 测点数据更新时候的回调函数
type DataPointsUpdatedFun func(points model.DataPoints, arg interface{}) interface{}

type notifyObject struct {
	Fun DataPointsUpdatedFun
	Arg interface{}
}

type notifyObjects []notifyObject

type rtdbModel struct {
	sync.Mutex
	data      MapData
	funMutex  sync.Mutex
	listeners notifyObjects
}

func init() {
	rtdbInstance.data = make(MapData)
}

// GetPv 获取测点值
func (r *rtdbModel) GetPv(id definition.DataPointIDType) (osal.Variant, bool) {
	r.Lock()
	defer r.Unlock()

	return r.getPv(id)
}

// GetVal 获取测点实时值
func (r *rtdbModel) GetVal(id definition.DataPointIDType) (model.RTValue, bool) {
	r.Lock()
	defer r.Unlock()

	return r.getVal(id, nil)
}

// SetVal 设置
func (r *rtdbModel) SetVal(id definition.DataPointIDType, val *model.RTValue, valChanged *bool) {
	r.Lock()
	defer r.Unlock()

	tmp := false
	if valChanged == nil {
		valChanged = &tmp
	} else {
		*valChanged = false
	}
	r.set(id, val, nil, nil, valChanged)
}

// GetAlarm 获取测点的告警状态
func (r *rtdbModel) GetAlarm(id definition.DataPointIDType) (model.RTAlarm, bool) {
	r.Lock()
	defer r.Unlock()

	if v, ok := r.data[id]; ok {
		return v.Alarm, true
	}
	return model.NewRTAlarm(), false
}

// SetAlarm 设置
func (r *rtdbModel) SetAlarm(id definition.DataPointIDType, alarm model.RTAlarm) {
	r.Lock()
	defer r.Unlock()

	if v, ok := r.data[id]; ok {
		v.Alarm = alarm
	} else {
		r.data[id] = &model.RTData{
			Val:   model.NewRTValue(),
			Alarm: alarm,
		}
	}
}

// GetDataPoints 获取测点数据
func (r *rtdbModel) GetDataPoints(points model.DataPoints) {
	r.Lock()
	defer r.Unlock()

	for i, point := range points {
		if v, ok := r.data[point.ID]; ok {
			points[i].Rtd = *v
		} else {
			points[i].Rtd = model.NewRTData()
		}
	}
}

// GetVirtualDataPoints 获取虚拟测点数据
func (r *rtdbModel) GetVirtualDataPoints(points model.DataPoints) {
	r.Lock()
	defer r.Unlock()

	for i, point := range points {
		if v, ok := r.data[point.ID]; ok {
			points[i].Rtd = *v
		} else {
			points[i].Rtd = model.NewVirtualRTData()
		}
	}
}

// SetDataPoints 写入实时数据库，
func (r *rtdbModel) SetDataPoints(points model.DataPoints, needNotify bool) {
	r.Lock()
	for i, point := range points {
		r.set(point.ID, &(points[i].Rtd.Val), nil, &(points[i].Rtd.Virtual), &(points[i].IsValueChanged))
	}
	r.Unlock()

	if needNotify {
		r.notify(points)
	}
}

// GetAll 获取所有测点数据
func (r *rtdbModel) GetAll() model.DataPoints {
	r.Lock()
	defer r.Unlock()

	points := make(model.DataPoints, len(r.data))
	i := 0
	for k, v := range r.data {
		points[i].ID = k
		points[i].Rtd = *v
		i++
	}
	return points
}

// ClearAll 清除所有测点数据
func (r *rtdbModel) ClearAll() {
	r.Lock()
	defer r.Unlock()

	r.data = make(MapData)
}

// ClearDataPoints 清除ids中的所有测点数据
func (r *rtdbModel) ClearDataPoints(ids definition.DataPointIDsType) {
	r.Lock()
	defer r.Unlock()

	for _, id := range ids {
		delete(r.data, id)
	}
}

func (r *rtdbModel) getPv(id definition.DataPointIDType) (osal.Variant, bool) {
	if v, ok := r.data[id]; ok {
		return v.Val.Pv, true
	}
	return osal.NewVariant(), false
}

func (r *rtdbModel) getVal(id definition.DataPointIDType, alarm *model.RTAlarm) (model.RTValue, bool) {
	if v, ok := r.data[id]; ok {
		if alarm != nil {
			*alarm = v.Alarm
		}
		return v.Val, true
	}
	return model.NewRTValue(), false
}

func (r *rtdbModel) set(id definition.DataPointIDType, val *model.RTValue, alarm *model.RTAlarm, virtual *bool, valChanged *bool) {
	*valChanged = false
	v, ok := r.data[id]
	if ok {
		if !v.Val.Pv.IsEqual(&val.Pv) || v.Val.Qua != val.Qua {
			*valChanged = true
		}
		if val.Qua != consts.QualityOk {
			// 如果 qua 异常时，不会更新已存在的 pv 值。
			v.Val.Qua = val.Qua
			v.Val.Desc = val.Desc
			v.Val.Tms = val.Tms
			// 并修改传入的 val.pv 为已存在的 pv 值
			val.Pv = v.Val.Pv
		} else {
			v.Val = *val
			v.Val.LastQuaOkTms = val.Tms
		}

		if alarm != nil {
			v.Alarm = *alarm
		}
		if virtual != nil {
			v.Virtual = *virtual
		}
	} else {
		rtd := &model.RTData{
			Val: *val,
		}
		if rtd.Val.Qua == consts.QualityOk {
			rtd.Val.LastQuaOkTms = rtd.Val.Tms
		}

		if alarm != nil {
			rtd.Alarm = *alarm
		} else {
			rtd.Alarm = model.NewRTAlarm()
		}

		if virtual != nil {
			rtd.Virtual = *virtual
		} else {
			rtd.Virtual = false
		}

		r.data[id] = rtd
		*valChanged = true
	}
}

// UnregisterDataPointsUpdatedFun 取消注册数据点更新回调函数
func (r *rtdbModel) UnregisterDataPointsUpdatedFun(fun DataPointsUpdatedFun) {
	if r == nil {
		return
	}
	r.funMutex.Lock()
	defer r.funMutex.Unlock()

	rhs := reflect.ValueOf(fun).Pointer()
	index := -1
	for i, listener := range r.listeners {
		lhs := reflect.ValueOf(listener.Fun).Pointer()
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

// RegisterDataPointsUpdatedFun 注册数据点更新回调函数
func (r *rtdbModel) RegisterDataPointsUpdatedFun(fun DataPointsUpdatedFun, arg interface{}) {
	if r == nil {
		return
	}
	r.funMutex.Lock()
	defer r.funMutex.Unlock()

	rhs := reflect.ValueOf(fun).Pointer()
	for _, listener := range r.listeners {
		lhs := reflect.ValueOf(listener.Fun).Pointer()
		if lhs == rhs {
			return
		}
	}
	r.listeners = append(r.listeners, notifyObject{
		Fun: fun,
		Arg: arg,
	})
}

func (r *rtdbModel) notify(points model.DataPoints) {
	r.funMutex.Lock()
	defer r.funMutex.Unlock()

	for _, listener := range r.listeners {
		listener.Fun(points, listener.Arg)
	}
}
