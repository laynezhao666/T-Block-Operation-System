// Package datatype 提供数据类型相关的定义和工具
package datatype

import "unsafe"

// DataType 数据类型枚举值
type DataType int

const (
	InvalidType DataType = iota
	BoolType    DataType = iota
	Int8Type    DataType = iota
	Int16Type   DataType = iota
	Int32Type   DataType = iota
	IntType     DataType = iota
	Int64Type   DataType = iota
	Uint8Type   DataType = iota
	Uint16Type  DataType = iota
	Uint32Type  DataType = iota
	UintType    DataType = iota
	Uint64Type  DataType = iota
	FloatType   DataType = iota
	DoubleType  DataType = iota
	StringType  DataType = iota
)

// GetDataTypeBytes 获取数据类型字节数
func GetDataTypeBytes(dataType DataType) int {
	switch dataType {
	case IntType, UintType:
		return int(unsafe.Sizeof(uint(0)))
	case BoolType, Int8Type, Uint8Type:
		return 1
	case Int16Type, Uint16Type:
		return 2
	case Int32Type, Uint32Type, FloatType:
		return 4
	case Int64Type, Uint64Type, DoubleType:
		return 8
	default:
		return -1
	}
}
