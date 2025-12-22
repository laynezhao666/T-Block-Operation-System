package utils

import (
	"agent/entity/definition/datatype"
	"strconv"
	"strings"
)

const (
	Int8Str   = "INT8"
	Int16Str  = "INT16"
	Int32Str  = "INT32"
	Int64Str  = "INT64"
	IntStr    = "INT"
	Uint8Str  = "UINT8"
	Uint16Str = "UINT16"
	Uint32Str = "UINT32"
	Uint64Str = "UINT64"
	UintStr   = "UINT"
	FloatStr  = "FLOAT"
	DoubleStr = "DOUBLE"
	BoolStr   = "BOOL"
	StringStr = "STRING"
)

// GetDataType 获取数据类型
func GetDataType(dataTypeString string, bitBegin *uint8, bitEnd *uint8) datatype.DataType {
	upperDataTypeString := strings.ToUpper(dataTypeString)
	switch upperDataTypeString {
	case Int8Str:
		return datatype.Int8Type
	case Int16Str:
		return datatype.Int16Type
	case Int32Str:
		return datatype.Int32Type
	case Int64Str:
		return datatype.Int64Type
	case IntStr:
		return datatype.IntType
	case Uint8Str:
		return datatype.Uint8Type
	case Uint16Str:
		return datatype.Uint16Type
	case Uint32Str:
		return datatype.Uint32Type
	case Uint64Str:
		return datatype.Uint64Type
	case UintStr:
		return datatype.UintType
	case FloatStr:
		return datatype.FloatType
	case DoubleStr:
		return datatype.DoubleType
	case StringStr:
		return datatype.StringType
	case BoolStr:
		return datatype.BoolType
	default:
		if strings.Index(upperDataTypeString, BoolStr) >= 0 {
			if bitBegin == nil || bitEnd == nil {
				break
			}
			if nums := strings.Split(upperDataTypeString, ":"); len(nums) == 2 { // bool0:1 形式
				if v, err := strconv.Atoi(strings.TrimLeft(nums[0], BoolStr)); err != nil {
					break
				} else {
					*bitBegin = uint8(v)
				}
				if v, err := strconv.Atoi(nums[1]); err != nil {
					break
				} else {
					*bitEnd = uint8(v)
				}
				return datatype.BoolType
			} else { //  bool1 形式
				if v, err := strconv.Atoi(strings.TrimLeft(nums[0], BoolStr)); err != nil {
					break
				} else {
					*bitBegin = uint8(v)
					*bitEnd = uint8(v + 1)
				}
				return datatype.BoolType
			}
		}
	}

	return datatype.InvalidType
}
