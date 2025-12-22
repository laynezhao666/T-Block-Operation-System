package osal

import (
	"errors"
	"fmt"
	"agent/utils"
	"math"
	"strconv"

	"agent/entity/definition"
	"agent/entity/definition/datatype"
)

var (
	ErrorVariantNil  = errors.New("variant is nil")
	ErrorInvalidType = errors.New("invalid type")
)

type Float = definition.FloatType

// Variant 用于存储数据
type Variant struct {
	dataType    datatype.DataType
	objectValue interface{}
}

// NewVariant 创建Variant
func NewVariant() Variant {
	return Variant{
		dataType:    datatype.InvalidType,
		objectValue: nil,
	}
}

// NewVariantWithValue 创建Variant
func NewVariantWithValue(value interface{}) Variant {
	v := Variant{}
	v.SetValue(value)
	return v
}

// String 用于打印
func (v *Variant) String() string {
	if v == nil {
		return ""
	}
	switch v.dataType {
	case datatype.InvalidType:
		return ""
	case datatype.FloatType:
		return strconv.FormatFloat(float64(v.objectValue.(float32)), 'f', -1, 32)
	case datatype.DoubleType:
		return strconv.FormatFloat(v.objectValue.(float64), 'f', -1, 64)
	}

	return fmt.Sprintf("%v", v.objectValue)
}

// AsBool 用于打印
func (v *Variant) AsBool() (bool, error) {
	if v == nil {
		return false, ErrorVariantNil
	}
	switch v.dataType {
	case datatype.InvalidType:
		return false, ErrorInvalidType
	case datatype.BoolType:
		if val, flag := v.objectValue.(bool); flag {
			return val, nil
		}
	case datatype.Int8Type:
		if val, flag := v.objectValue.(int8); flag {
			return val != 0, nil
		}
	case datatype.Int16Type:
		if val, flag := v.objectValue.(int16); flag {
			return val != 0, nil
		}
	case datatype.Int32Type:
		if val, flag := v.objectValue.(int32); flag {
			return val != 0, nil
		}
	case datatype.IntType:
		if val, flag := v.objectValue.(int); flag {
			return val != 0, nil
		}
	case datatype.Int64Type:
		if val, flag := v.objectValue.(int64); flag {
			return val != 0, nil
		}
	case datatype.Uint8Type:
		if val, flag := v.objectValue.(uint8); flag {
			return val != 0, nil
		}
	case datatype.Uint16Type:
		if val, flag := v.objectValue.(uint16); flag {
			return val != 0, nil
		}
	case datatype.Uint32Type:
		if val, flag := v.objectValue.(uint32); flag {
			return val != 0, nil
		}
	case datatype.UintType:
		if val, flag := v.objectValue.(uint); flag {
			return val != 0, nil
		}
	case datatype.Uint64Type:
		if val, flag := v.objectValue.(uint64); flag {
			return val != 0, nil
		}
	case datatype.FloatType:
		if val, flag := v.objectValue.(float32); flag {
			return !utils.IsFloat32Zero(val), nil
		}
	case datatype.DoubleType:
		if val, flag := v.objectValue.(float64); flag {
			return !utils.IsFloat64Zero(val), nil
		}
	case datatype.StringType:
		if val, flag := v.objectValue.(string); flag {
			f, err := strconv.ParseFloat(val, 64)
			if err != nil {
				break
			}
			return !utils.IsFloat64Zero(f), nil
		}
	default:
		return false, ErrorInvalidType
	}
	return false, fmt.Errorf("%+v: type assertion failed", v.objectValue)
}

// AsString 转换为string
func (v *Variant) AsString() (string, error) {
	if v == nil {
		return "", ErrorVariantNil
	}

	switch v.dataType {
	case datatype.BoolType:
		// 采集层直采数据，不存在true、false，均转换为1，0
		if val, flag := v.objectValue.(bool); flag {
			vs := "0"
			if val {
				vs = "1"
			}
			v.SetValue(vs)
			return vs, nil
		}
	default:
		return v.String(), nil
	}
	return v.String(), nil
}

// AsDouble 转换为float64
func (v *Variant) AsDouble() (float64, error) {
	if v == nil {
		return -1.0, ErrorVariantNil
	}
	switch v.dataType {
	case datatype.InvalidType:
		return -1.0, ErrorInvalidType
	case datatype.BoolType:
		if val, flag := v.objectValue.(bool); flag {
			if val {
				return 1.0, nil
			} else {
				return 0.0, nil
			}
		}
	case datatype.Int8Type:
		if val, flag := v.objectValue.(int8); flag {
			return float64(val), nil
		}
	case datatype.Int16Type:
		if val, flag := v.objectValue.(int16); flag {
			return float64(val), nil
		}
	case datatype.Int32Type:
		if val, flag := v.objectValue.(int32); flag {
			return float64(val), nil
		}
	case datatype.IntType:
		if val, flag := v.objectValue.(int); flag {
			return float64(val), nil
		}
	case datatype.Int64Type:
		if val, flag := v.objectValue.(int64); flag {
			return float64(val), nil
		}
	case datatype.Uint8Type:
		if val, flag := v.objectValue.(uint8); flag {
			return float64(val), nil
		}
	case datatype.Uint16Type:
		if val, flag := v.objectValue.(uint16); flag {
			return float64(val), nil
		}
	case datatype.Uint32Type:
		if val, flag := v.objectValue.(uint32); flag {
			return float64(val), nil
		}
	case datatype.UintType:
		if val, flag := v.objectValue.(uint); flag {
			return float64(val), nil
		}
	case datatype.Uint64Type:
		if val, flag := v.objectValue.(uint64); flag {
			return float64(val), nil
		}
	case datatype.FloatType:
		if val, flag := v.objectValue.(float32); flag {
			return float64(val), nil
		}
	case datatype.DoubleType:
		if val, flag := v.objectValue.(float64); flag {
			return val, nil
		}
	case datatype.StringType:
		if val, flag := v.objectValue.(string); flag {
			f, err := strconv.ParseFloat(val, 64)
			if err != nil {
				break
			}
			return f, nil
		}
	default:
		return -1.0, ErrorInvalidType
	}
	return -1.0, fmt.Errorf("%+v: type assertion failed", v.objectValue)
}

// AsFloat AsFloat
func (v *Variant) AsFloat() (Float, error) {
	if v == nil {
		return Float(-1.0), ErrorVariantNil
	}
	switch v.dataType {
	case datatype.InvalidType:
		return Float(-1.0), ErrorInvalidType
	case datatype.BoolType:
		if val, flag := v.objectValue.(bool); flag {
			if val {
				return Float(1.0), nil
			} else {
				return Float(0.0), nil
			}
		} else { // 尝试按照int解析，兼容原版simulator模拟数据
			if val, flag := v.objectValue.(int); flag {
				if val == 1 || val == 0 {
					return Float(val), nil
				}
			}
		}
	case datatype.Int8Type:
		if val, flag := v.objectValue.(int8); flag {
			return Float(val), nil
		}
	case datatype.Int16Type:
		if val, flag := v.objectValue.(int16); flag {
			return Float(val), nil
		}
	case datatype.Int32Type:
		if val, flag := v.objectValue.(int32); flag {
			return Float(val), nil
		}
	case datatype.IntType:
		if val, flag := v.objectValue.(int); flag {
			return Float(val), nil
		}
	case datatype.Int64Type:
		if val, flag := v.objectValue.(int64); flag {
			return Float(val), nil
		}
	case datatype.Uint8Type:
		if val, flag := v.objectValue.(uint8); flag {
			return Float(val), nil
		}
	case datatype.Uint16Type:
		if val, flag := v.objectValue.(uint16); flag {
			return Float(val), nil
		}
	case datatype.Uint32Type:
		if val, flag := v.objectValue.(uint32); flag {
			return Float(val), nil
		}
	case datatype.UintType:
		if val, flag := v.objectValue.(uint); flag {
			return Float(val), nil
		}
	case datatype.Uint64Type:
		if val, flag := v.objectValue.(uint64); flag {
			return Float(val), nil
		}
	case datatype.FloatType:
		if val, flag := v.objectValue.(float32); flag {
			return Float(val), nil
		}
	case datatype.DoubleType:
		if val, flag := v.objectValue.(float64); flag {
			return Float(val), nil
		}
	case datatype.StringType:
		if val, flag := v.objectValue.(string); flag {
			f, err := strconv.ParseFloat(val, 64)
			if err != nil {
				break
			}
			return Float(f), nil
		}
	default:
		return Float(-1.0), ErrorInvalidType
	}
	return Float(-1.0), fmt.Errorf("%+v: type assertion failed", v.objectValue)
}

// AsInt64 as int64
func (v *Variant) AsInt64() (int64, error) {
	if v == nil {
		return -1, ErrorVariantNil
	}
	switch v.dataType {
	case datatype.InvalidType:
		return -1, ErrorInvalidType
	case datatype.BoolType:
		if val, flag := v.objectValue.(bool); flag {
			if val {
				return 1, nil
			} else {
				return 0, nil
			}
		}
	case datatype.Int8Type:
		if val, flag := v.objectValue.(int8); flag {
			return int64(val), nil
		}
	case datatype.Int16Type:
		if val, flag := v.objectValue.(int16); flag {
			return int64(val), nil
		}
	case datatype.Int32Type:
		if val, flag := v.objectValue.(int32); flag {
			return int64(val), nil
		}
	case datatype.IntType:
		if val, flag := v.objectValue.(int); flag {
			return int64(val), nil
		}
	case datatype.Int64Type:
		if val, flag := v.objectValue.(int64); flag {
			return val, nil
		}
	case datatype.Uint8Type:
		if val, flag := v.objectValue.(uint8); flag {
			return int64(val), nil
		}
	case datatype.Uint16Type:
		if val, flag := v.objectValue.(uint16); flag {
			return int64(val), nil
		}
	case datatype.Uint32Type:
		if val, flag := v.objectValue.(uint32); flag {
			return int64(val), nil
		}
	case datatype.UintType:
		if val, flag := v.objectValue.(uint); flag {
			return int64(val), nil
		}
	case datatype.Uint64Type:
		if val, flag := v.objectValue.(uint64); flag {
			return int64(val), nil
		}
	case datatype.FloatType:
		if val, flag := v.objectValue.(float32); flag {
			return int64(val), nil
		}
	case datatype.DoubleType:
		if val, flag := v.objectValue.(float64); flag {
			return int64(val), nil
		}
	case datatype.StringType:
		if val, flag := v.objectValue.(string); flag {
			f, err := strconv.ParseInt(val, 0, 64)
			if err != nil {
				break
			}
			return f, nil
		}
	default:
		return -1, ErrorInvalidType
	}
	return -1, fmt.Errorf("%+v: type assertion failed", v.objectValue)
}

// AsUint64 AsUint64
func (v *Variant) AsUint64() (uint64, error) {
	if v == nil {
		return math.MaxUint64, ErrorVariantNil
	}
	switch v.dataType {
	case datatype.InvalidType:
		return math.MaxUint64, ErrorInvalidType
	case datatype.BoolType:
		if val, flag := v.objectValue.(bool); flag {
			if val {
				return 1, nil
			} else {
				return 0, nil
			}
		}
	case datatype.Int8Type:
		if val, flag := v.objectValue.(int8); flag {
			return uint64(val), nil
		}
	case datatype.Int16Type:
		if val, flag := v.objectValue.(int16); flag {
			return uint64(val), nil
		}
	case datatype.Int32Type:
		if val, flag := v.objectValue.(int32); flag {
			return uint64(val), nil
		}
	case datatype.IntType:
		if val, flag := v.objectValue.(int); flag {
			return uint64(val), nil
		}
	case datatype.Int64Type:
		if val, flag := v.objectValue.(int64); flag {
			return uint64(val), nil
		}
	case datatype.Uint8Type:
		if val, flag := v.objectValue.(uint8); flag {
			return uint64(val), nil
		}
	case datatype.Uint16Type:
		if val, flag := v.objectValue.(uint16); flag {
			return uint64(val), nil
		}
	case datatype.Uint32Type:
		if val, flag := v.objectValue.(uint32); flag {
			return uint64(val), nil
		}
	case datatype.UintType:
		if val, flag := v.objectValue.(uint); flag {
			return uint64(val), nil
		}
	case datatype.Uint64Type:
		if val, flag := v.objectValue.(uint64); flag {
			return val, nil
		}
	case datatype.FloatType:
		if val, flag := v.objectValue.(float32); flag {
			return uint64(val), nil
		}
	case datatype.DoubleType:
		if val, flag := v.objectValue.(float64); flag {
			return uint64(val), nil
		}
	case datatype.StringType:
		if val, flag := v.objectValue.(string); flag {
			f, err := strconv.ParseUint(val, 0, 64)
			if err != nil {
				break
			}
			return f, nil
		}
	default:
		return math.MaxUint64, ErrorInvalidType
	}
	return math.MaxUint64, fmt.Errorf("%+v: type assertion failed", v.objectValue)
}

// SetType 设置数据类型
func (v *Variant) SetType(t datatype.DataType) {
	if v == nil {
		return
	}

	v.dataType = t
}

// SetFloat64 设置float64类型
func (v *Variant) SetFloat64(value float64) {
	if v == nil {
		return
	}

	switch v.dataType {
	case datatype.InvalidType:
		return
	case datatype.BoolType:
		if utils.IsFloat64Zero(value) {
			v.objectValue = 0
		} else {
			v.objectValue = 1
		}
		return
	case datatype.Int8Type:
		v.objectValue = int8(value)
		return
	case datatype.Int16Type:
		v.objectValue = int16(value)
		return
	case datatype.Int32Type:
		v.objectValue = int32(value)
		return
	case datatype.IntType:
		v.objectValue = int(value)
		return
	case datatype.Int64Type:
		v.objectValue = int64(value)
		return
	case datatype.Uint8Type:
		v.objectValue = uint8(value)
		return
	case datatype.Uint16Type:
		v.objectValue = uint16(value)
		return
	case datatype.Uint32Type:
		v.objectValue = uint32(value)
		return
	case datatype.UintType:
		v.objectValue = uint(value)
		return
	case datatype.Uint64Type:
		v.objectValue = uint64(value)
		return
	case datatype.FloatType:
		v.objectValue = float32(value)
		return
	case datatype.DoubleType:
		v.objectValue = value
		return
	case datatype.StringType:
		v.objectValue = fmt.Sprintf("%v", value)
		return
	default:
		return
	}
}

// SetValue 设置值
func (v *Variant) SetValue(value interface{}) {
	if v == nil {
		return
	}
	v.objectValue = value

	switch value.(type) {
	case nil:
		v.dataType = datatype.InvalidType
	case bool:
		v.dataType = datatype.BoolType
	case int8:
		v.dataType = datatype.Int8Type
	case int16:
		v.dataType = datatype.Int16Type
	case int32:
		v.dataType = datatype.Int32Type
	case int:
		v.dataType = datatype.IntType
	case int64:
		v.dataType = datatype.Int64Type
	case uint8:
		v.dataType = datatype.Uint8Type
	case uint16:
		v.dataType = datatype.Uint16Type
	case uint32:
		v.dataType = datatype.Uint32Type
	case uint:
		v.dataType = datatype.UintType
	case uint64:
		v.dataType = datatype.Uint64Type
	case float32:
		v.dataType = datatype.FloatType
	case float64:
		v.dataType = datatype.DoubleType
	case string:
		v.dataType = datatype.StringType
	default:
		v.dataType = datatype.InvalidType
	}
}

// IsComparable 判断两者是否可比，数字与数字可比，字符串与字符串可比，数字与字符串不可比
func (v *Variant) IsComparable(rhs *Variant) bool {
	if v == nil || rhs == nil {
		return false
	}
	if v.dataType == datatype.InvalidType || rhs.dataType == datatype.InvalidType {
		return false
	}
	if v.dataType == datatype.StringType {
		return rhs.dataType == datatype.StringType
	}
	return true
}

// IsEqual 判断两个 Variant 对象是否相等
func (v *Variant) IsEqual(rhs *Variant) bool {
	if v == nil {
		return rhs == nil
	}
	// 两者都是invalidType，认为相等,即无需触发变化上报
	if v.dataType == datatype.InvalidType && rhs.dataType == datatype.InvalidType {
		return true
	}
	if !v.IsComparable(rhs) {
		return false
	}
	switch v.dataType {
	case datatype.StringType:
		l, ok1 := v.objectValue.(string)
		r, ok2 := rhs.objectValue.(string)
		return ok1 && ok2 && l == r
	case datatype.FloatType:
		l, ok := v.objectValue.(float32)
		r, err := rhs.AsDouble()
		return ok && err == nil && utils.IsFloat64Equal(float64(l), r)
	case datatype.DoubleType:
		l, ok := v.objectValue.(float64)
		r, err := rhs.AsDouble()
		return ok && err == nil && utils.IsFloat64Equal(l, r)
	case datatype.BoolType:
		l, ok := v.objectValue.(bool)
		r, err := rhs.AsBool()
		return ok && err == nil && l == r
	case datatype.Int8Type:
		l, ok := v.objectValue.(int8)
		r, err := rhs.AsInt64()
		return ok && err == nil && int64(l) == r
	case datatype.Int16Type:
		l, ok := v.objectValue.(int16)
		r, err := rhs.AsInt64()
		return ok && err == nil && int64(l) == r
	case datatype.Int32Type:
		l, ok := v.objectValue.(int32)
		r, err := rhs.AsInt64()
		return ok && err == nil && int64(l) == r
	case datatype.IntType:
		l, ok := v.objectValue.(int)
		r, err := rhs.AsInt64()
		return ok && err == nil && int64(l) == r
	case datatype.Int64Type:
		l, ok := v.objectValue.(int64)
		r, err := rhs.AsInt64()
		return ok && err == nil && l == r
	case datatype.Uint8Type:
		l, ok := v.objectValue.(uint8)
		r, err := rhs.AsUint64()
		return ok && err == nil && uint64(l) == r
	case datatype.Uint16Type:
		l, ok := v.objectValue.(uint16)
		r, err := rhs.AsUint64()
		return ok && err == nil && uint64(l) == r
	case datatype.Uint32Type:
		l, ok := v.objectValue.(uint32)
		r, err := rhs.AsUint64()
		return ok && err == nil && uint64(l) == r
	case datatype.UintType:
		l, ok := v.objectValue.(uint)
		r, err := rhs.AsUint64()
		return ok && err == nil && uint64(l) == r
	case datatype.Uint64Type:
		l, ok := v.objectValue.(uint64)
		r, err := rhs.AsUint64()
		return ok && err == nil && l == r
	default:
		// 其他情况，返回 false
	}
	return false
}

// IsZero 判断是否为0
func (v *Variant) IsZero() (bool, error) {
	if v == nil {
		return false, errors.New("nil")
	}
	switch v.dataType {
	case datatype.BoolType:
		return v.isBoolZero()
	case datatype.Int8Type:
		return v.isInt8Zero()
	case datatype.Int16Type:
		return v.isInt16Zero()
	case datatype.Int32Type:
		return v.isInt32Zero()
	case datatype.IntType:
		return v.isIntZero()
	case datatype.Int64Type:
		return v.isInt64Zero()
	case datatype.Uint8Type:
		return v.isUint8Zero()
	case datatype.Uint16Type:
		return v.isUint16Zero()
	case datatype.Uint32Type:
		return v.isUint32Zero()
	case datatype.UintType:
		return v.isUintZero()
	case datatype.Uint64Type:
		return v.isUint64Zero()
	case datatype.FloatType:
		return v.isFloat32Zero()
	case datatype.DoubleType:
		return v.isFloat64Zero()
	case datatype.StringType:
		return v.isStringZero()
	default:
		return false, fmt.Errorf("invalid type or value, type: %v, value: %+v", v.dataType, v.objectValue)
	}
}

func (v *Variant) isBoolZero() (bool, error) {
	x, ok := v.objectValue.(bool)
	if !ok {
		return false, fmt.Errorf("invalid bool value: %+v", v.objectValue)
	}
	return x == false, nil
}

func (v *Variant) isInt8Zero() (bool, error) {
	x, ok := v.objectValue.(int8)
	if !ok {
		return false, fmt.Errorf("invalid int8 value: %+v", v.objectValue)
	}
	return x == 0, nil
}

func (v *Variant) isInt16Zero() (bool, error) {
	x, ok := v.objectValue.(int16)
	if !ok {
		return false, fmt.Errorf("invalid int16 value: %+v", v.objectValue)
	}
	return x == 0, nil
}

func (v *Variant) isInt32Zero() (bool, error) {
	x, ok := v.objectValue.(int32)
	if !ok {
		return false, fmt.Errorf("invalid int32 value: %+v", v.objectValue)
	}
	return x == 0, nil
}

func (v *Variant) isIntZero() (bool, error) {
	x, ok := v.objectValue.(int)
	if !ok {
		return false, fmt.Errorf("invalid int value: %+v", v.objectValue)
	}
	return x == 0, nil
}

func (v *Variant) isInt64Zero() (bool, error) {
	x, ok := v.objectValue.(int64)
	if !ok {
		return false, fmt.Errorf("invalid int64 value: %+v", v.objectValue)
	}
	return x == 0, nil
}

func (v *Variant) isUint8Zero() (bool, error) {
	x, ok := v.objectValue.(uint8)
	if !ok {
		return false, fmt.Errorf("invalid uint8 value: %+v", v.objectValue)
	}
	return x == 0, nil
}

func (v *Variant) isUint16Zero() (bool, error) {
	x, ok := v.objectValue.(uint16)
	if !ok {
		return false, fmt.Errorf("invalid uint16 value: %+v", v.objectValue)
	}
	return x == 0, nil
}

func (v *Variant) isUint32Zero() (bool, error) {
	x, ok := v.objectValue.(uint32)
	if !ok {
		return false, fmt.Errorf("invalid uint32 value: %+v", v.objectValue)
	}
	return x == 0, nil
}

func (v *Variant) isUintZero() (bool, error) {
	x, ok := v.objectValue.(uint)
	if !ok {
		return false, fmt.Errorf("invalid uint value: %+v", v.objectValue)
	}
	return x == 0, nil
}

func (v *Variant) isUint64Zero() (bool, error) {
	x, ok := v.objectValue.(uint64)
	if !ok {
		return false, fmt.Errorf("invalid uint64 value: %+v", v.objectValue)
	}
	return x == 0, nil
}

func (v *Variant) isFloat32Zero() (bool, error) {
	x, ok := v.objectValue.(float32)
	if !ok {
		return false, fmt.Errorf("invalid float32 value: %+v", v.objectValue)
	}
	return utils.IsFloat32Zero(x), nil
}

func (v *Variant) isFloat64Zero() (bool, error) {
	x, ok := v.objectValue.(float64)
	if !ok {
		return false, fmt.Errorf("invalid float64 value: %+v", v.objectValue)
	}
	return utils.IsFloat64Zero(x), nil
}

func (v *Variant) isStringZero() (bool, error) {
	x, ok := v.objectValue.(string)
	if !ok {
		return false, fmt.Errorf("invalid string value: %+v", v.objectValue)
	}
	return x == "0", nil
}
