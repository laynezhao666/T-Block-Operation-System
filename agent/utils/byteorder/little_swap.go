package byteorder

import (
	"math"
)

var LittleEndianSwapExtend littleEndianSwap

type littleEndianSwap struct{}

// Uint16 little endian
func (littleEndianSwap) Uint16(b []byte) uint16 {
	_ = b[1]
	return uint16(b[1]) | uint16(b[0])<<8
}

// Uint32 little endian
func (littleEndianSwap) Uint32(b []byte) uint32 {
	_ = b[3]
	return uint32(b[1]) | uint32(b[0])<<8 | uint32(b[3])<<16 | uint32(b[2])<<24
}

// Uint64 little endian
func (le littleEndianSwap) Uint64(b []byte) uint64 {
	low32 := le.Uint32(b[:4])
	high32 := le.Uint32(b[4:])
	return uint64(high32)<<32 | uint64(low32)
}

// Float little endian
func (le littleEndianSwap) Float(b []byte) float32 {
	return math.Float32frombits(le.Uint32(b))
}

// Double little endian
func (le littleEndianSwap) Double(b []byte) float64 {
	return math.Float64frombits(le.Uint64(b))
}

// PutUint16 little endian
func (littleEndianSwap) PutUint16(b []byte, v uint16) {
	// TODO, 当前未使用
}

// PutUint32 little endian
func (littleEndianSwap) PutUint32(b []byte, v uint32) {
	// TODO, 当前未使用
}

// PutUint64 little endian
func (littleEndianSwap) PutUint64(b []byte, v uint64) {
	// TODO, 当前未使用
}

// PutFloat little endian
func (littleEndianSwap) PutFloat(b []byte, v float32) {
	// TODO, 当前未使用
}

// PutDouble little endian
func (littleEndianSwap) PutDouble(b []byte, v float64) {
	// TODO, 当前未使用
}

// String little endian
func (littleEndianSwap) String() string {
	return "LittleEndianSwap"
}
