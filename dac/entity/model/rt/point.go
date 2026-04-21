// Package rt 定义门禁系统的实时数据模型。
package rt

import (
	"fmt"
	"time"

	"dac/entity/consts"
)

// RTValue 测点实时值
type RTValue struct {
	// 测点值
	Pv string `json:"pv"`
	// 测点质量
	Qua consts.Quality `json:"qua"`
	// 采集时间，单位：毫秒
	Timestamp int64 `json:"tms"`
}

// NewRTValueWithPvTime 创建带值和时间的实时值
func NewRTValueWithPvTime(value string, t time.Time) RTValue {
	return RTValue{
		Pv:        value,
		Qua:       consts.QualityOK,
		Timestamp: t.UnixMilli(),
	}
}

// NewRTValue 创建空的实时值（质量为不确定）
func NewRTValue() RTValue {
	return RTValue{
		Pv:        "",
		Qua:       consts.QualityUncertain,
		Timestamp: 0,
	}
}

// Point 测点数据
type Point struct {
	// 测点 ID
	ID string `json:"id"`
	// 测点数据
	Rtd            RTValue `json:"rtd"`
	IsValueChanged bool
}

// NewPoint 创建指定ID的测点实例
func NewPoint(id string) Point {
	return Point{
		ID:             id,
		Rtd:            NewRTValue(),
		IsValueChanged: false,
	}
}

// SetTime 设置测点的采集时间戳（毫秒）
func (p *Point) SetTime(tms int64) *Point {
	if p == nil {
		return nil
	}

	p.Rtd.Timestamp = tms
	return p
}

// SetValue 设置测点值（自动转换为字符串）
func (p *Point) SetValue(value interface{}) *Point {
	if p == nil {
		return nil
	}

	if v, ok := value.(string); ok {
		p.Rtd.Pv = v
	} else {
		p.Rtd.Pv = fmt.Sprintf("%v", value)
	}

	return p
}

// SetValueWithTime 同时设置测点值和时间戳，质量标记为OK
func (p *Point) SetValueWithTime(
	value interface{}, tms int64,
) *Point {
	return p.SetValue(value).SetTime(tms).SetQua(consts.QualityOK)
}

// SetQua 设置测点质量标识
func (p *Point) SetQua(qua consts.Quality) *Point {
	if p == nil {
		return nil
	}

	p.Rtd.Qua = qua
	return p
}

// Points 测点列表类型
type Points = []Point
