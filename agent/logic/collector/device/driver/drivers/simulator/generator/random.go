package generator

import (
	"math/rand"
)

// RandomImpl 随机数生成器
type RandomImpl struct {
	low float64
	up  float64
	len float64
}

// NewRandomImpl 创建随机数生成器
func NewRandomImpl(low, up float64) *RandomImpl {
	if low > up {
		low, up = up, low
	}
	return &RandomImpl{
		low: low,
		up:  up,
		len: up - low,
	}
}

// Get 获取随机数
func (r *RandomImpl) Get() float64 {
	return rand.Float64()*r.len + r.low
}
