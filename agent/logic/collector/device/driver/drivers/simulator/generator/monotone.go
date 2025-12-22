package generator

import (
	"agent/utils"
	"math"
)

// MonotoneImpl 单调递增
type MonotoneImpl struct {
	low   float64
	up    float64
	step  float64
	len   float64
	equal bool
	x     float64
}

// NewMonotoneImpl 创建单调递增数据生成器
func NewMonotoneImpl(low, up, step float64) *MonotoneImpl {
	if low > up {
		low, up = up, low
	}
	l := up - low
	equal := utils.IsFloat64Zero(l)
	if equal {
		step = 0
	} else {
		step = math.Mod(step, l)
	}
	return &MonotoneImpl{
		low:   low,
		up:    up,
		step:  step,
		len:   l,
		equal: equal,
		x:     low,
	}
}

// Get 获取数据
func (m *MonotoneImpl) Get() float64 {
	if m.equal {
		return m.low
	}

	m.x += m.step
	if m.x > m.up {
		m.x -= m.len
	}
	if m.x < m.low {
		m.x += m.len
	}
	return m.x
}
