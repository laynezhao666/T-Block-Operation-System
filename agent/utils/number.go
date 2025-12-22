package utils

import (
	"math"
)

const (
	Epsilon = 1.0e-6
)

// IsSpecialValue 判断是否为特殊值
func IsSpecialValue(v float32) bool {
	// -9999 上报无法支持, 北向接口中有定义但由于客观原因无法支持的测点
	if v == float32(-9999) {
		return true
	}

	// -99999 上报未初始化,本地监控系统启动时，测点的初始值上报
	if v == float32(-99999) {
		return true
	}

	// 上报通讯中断, 或者设备通讯中断时，测点上报-99998
	// 测点未在北向接口中定义，上报-99997
	if v >= float32(-99998) && v <= float32(-99990) {
		return true
	}

	return false
}

// IsFloat64Equal 判断两个浮点数是否相等
func IsFloat64Equal(lhs float64, rhs float64) bool {
	return math.Abs(lhs-rhs) < Epsilon
}

// IsFloat32Equal 判断两个浮点数是否相等
func IsFloat32Equal(lhs float32, rhs float32) bool {
	return math.Abs(float64(lhs-rhs)) < Epsilon
}

// IsFloat32Zero 判断浮点数是否为0
func IsFloat32Zero(x float32) bool {
	return IsFloat32Equal(x, 0.0)
}

// IsFloat64Zero 判断浮点数是否为0
func IsFloat64Zero(x float64) bool {
	return IsFloat64Equal(x, 0.0)
}

// Bool2Int bool转int
func Bool2Int(b bool) int {
	if b {
		return 1
	}
	return 0
}

// SubtractUint64 减法
func SubtractUint64(a, b uint64) uint64 {
	if a >= b {
		return a - b
	}

	return math.MaxUint64 - (b - a) + 1
}
