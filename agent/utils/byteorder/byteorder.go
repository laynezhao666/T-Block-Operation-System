// Package byteorder provides utilities for handling byte order (endianness)
package byteorder

import (
	"encoding/binary"
	"unsafe"
)

type ByteOrderType int

const (
	ByteOrderBig        ByteOrderType = iota
	ByteOrderLittle     ByteOrderType = iota
	ByteOrderBigSwap    ByteOrderType = iota
	ByteOrderLittleSwap ByteOrderType = iota
)

var (
	NativeByteOrder ByteOrderType
)

// ByteOrderExtend 扩展的字节序
type ByteOrderExtend interface {
	// Float 读取float
	Float(b []byte) float32
	// Double 读取double
	Double(b []byte) float64
	// PutFloat 写入float
	PutFloat(b []byte, v float32)
	// PutDouble 写入double
	PutDouble(b []byte, v float64)
	binary.ByteOrder
}

func init() {
	temp := 0xaabb
	if *(*byte)(unsafe.Pointer(&temp)) == 0xaa {
		NativeByteOrder = ByteOrderBig
	} else {
		NativeByteOrder = ByteOrderLittle
	}
}

// CopyValue 复制字节序
func CopyValue(dst []byte, dstByteOrder ByteOrderType, src []byte, srcByteOrder ByteOrderType) bool {
	if dstByteOrder == srcByteOrder {
		copy(dst, src)
		return true
	}
	return false
}

// CopyValueToNative 复制字节序到本地字节序
func CopyValueToNative(dst, src []byte, srcByteOrder ByteOrderType) bool {
	return CopyValue(dst, NativeByteOrder, src, srcByteOrder)
}
