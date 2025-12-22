package utils

import "math"

const (
	kAbs = "abs"
	kNot = "not"
)

// AbsInt Abs函数处理各种数值类型的绝对值
func AbsInt(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

// AbsInt64 Abs函数处理各种数值类型的绝对值
func AbsInt64(n int64) int64 {
	if n < 0 {
		return -n
	}
	return n
}

// AbsUint16 Abs函数处理各种数值类型的绝对值
func AbsUint16(n uint16) uint16 { return n }

// AbsUint32 Abs函数处理各种数值类型的绝对值
func AbsUint32(n uint32) uint32 { return n }

// AbsUint64 Abs函数处理各种数值类型的绝对值
func AbsUint64(n uint64) uint64 { return n }

// AbsFloat32 Abs函数处理各种数值类型的绝对值
func AbsFloat32(n float32) float32 {
	return float32(math.Abs(float64(n)))
}

// AbsFloat64 Abs函数处理各种数值类型的绝对值
func AbsFloat64(n float64) float64 {
	return math.Abs(n)
}

// AbsBool Abs函数处理布尔值的绝对值
func AbsBool(b bool) bool {
	return b // 布尔值的绝对值定义为原值
}

// Unary Unary函数根据操作符执行abs或not操作
func Unary(op string, value interface{}) interface{} {
	if op == "" {
		return value
	}
	switch op {
	case kAbs:
		return absValue(value)
	case kNot:
		return notValue(value)
	default:
		return value
	}
}

// 处理绝对值
func absValue(value interface{}) interface{} {
	switch v := value.(type) {
	case int:
		return AbsInt(v)
	case int64:
		return AbsInt64(v)
	case uint16:
		return AbsUint16(v)
	case uint32:
		return AbsUint32(v)
	case uint64:
		return AbsUint64(v)
	case float32:
		return AbsFloat32(v)
	case float64:
		return AbsFloat64(v)
	case bool:
		return AbsBool(v)
	default:
		return value // 未知类型返回原值
	}
}

// 处理逻辑非
func notValue(value interface{}) interface{} {
	switch v := value.(type) {
	case bool:
		return !v
	case int:
		if v != 0 {
			return 0
		}
		return 1
	case int64:
		if v != 0 {
			return int64(0)
		}
		return int64(1)
	case uint16:
		if v != 0 {
			return uint16(0)
		}
		return uint16(1)
	case uint32:
		if v != 0 {
			return uint32(0)
		}
		return uint32(1)
	case uint64:
		if v != 0 {
			return uint64(0)
		}
		return uint64(1)
	case float32:
		if v != 0 {
			return float32(0)
		}
		return float32(1)
	case float64:
		if v != 0 {
			return 0.0
		}
		return 1.0
	default:
		return value // 未知类型返回原值
	}
}

// IsUnaryFun 判断是否为支持的一元操作符
func IsUnaryFun(op string) bool {
	return op == kAbs || op == kNot
}
