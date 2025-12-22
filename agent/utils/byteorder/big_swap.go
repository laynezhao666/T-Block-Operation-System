package byteorder

import (
	"math"
)

var BigEndianSwapExtend bigEndianSwap

type bigEndianSwap struct{}

// Uint16 returns the uint16 encoded in the first 2 bytes of b.
func (bigEndianSwap) Uint16(b []byte) uint16 {
	_ = b[1]
	return uint16(b[0]) | uint16(b[1])<<8
}

// Uint32 returns the uint32 encoded in the first 4 bytes of b.
func (bigEndianSwap) Uint32(b []byte) uint32 {
	_ = b[3]
	return uint32(b[2]) | uint32(b[3])<<8 | uint32(b[0])<<16 | uint32(b[1])<<24
}

// Uint64 returns the uint64 encoded in the first 8 bytes of b.
func (be bigEndianSwap) Uint64(b []byte) uint64 {
	low32 := be.Uint32(b[4:])
	high32 := be.Uint32(b[:4])
	return uint64(high32)<<32 | uint64(low32)
}

// Float returns the float32 encoded in the first 4 bytes of b.
func (be bigEndianSwap) Float(b []byte) float32 {
	return math.Float32frombits(be.Uint32(b))
}

// Double returns the float64 encoded in the first 8 bytes of b.
func (be bigEndianSwap) Double(b []byte) float64 {
	return math.Float64frombits(be.Uint64(b))
}

// PutUint16 stores v into b.
func (bigEndianSwap) PutUint16(b []byte, v uint16) {
	// TODO, 当前未使用
}

// PutUint32 stores v into b.
func (bigEndianSwap) PutUint32(b []byte, v uint32) {
	// TODO, 当前未使用
}

// PutUint64 stores v into b.
func (bigEndianSwap) PutUint64(b []byte, v uint64) {
	// TODO, 当前未使用
}

// PutFloat stores v into b.
func (bigEndianSwap) PutFloat(b []byte, v float32) {
	// TODO, 当前未使用
}

// PutDouble stores v into b.
func (bigEndianSwap) PutDouble(b []byte, v float64) {
	// TODO, 当前未使用
}

// String returns the string representation of the byte order.
func (bigEndianSwap) String() string {
	return "BigEndianSwap"
}
