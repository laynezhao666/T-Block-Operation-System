package byteorder

import (
	"encoding/binary"
	"math"
)

var LittleEndianExtend littleEndian

type littleEndian struct{}

// Uint16 returns the uint16 encoded in b.
func (littleEndian) Uint16(b []byte) uint16 {
	return binary.LittleEndian.Uint16(b)
}

// Uint32 returns the uint32 encoded in b.
func (littleEndian) Uint32(b []byte) uint32 {
	return binary.LittleEndian.Uint32(b)
}

// Uint64 returns the uint64 encoded in b.
func (littleEndian) Uint64(b []byte) uint64 {
	return binary.LittleEndian.Uint64(b)
}

// Float returns the float32 encoded in b.
func (be littleEndian) Float(b []byte) float32 {
	return math.Float32frombits(be.Uint32(b))
}

// Double returns the float64 encoded in b.
func (be littleEndian) Double(b []byte) float64 {
	return math.Float64frombits(be.Uint64(b))
}

// PutUint16 encodes v into b.
func (littleEndian) PutUint16(b []byte, v uint16) {
	binary.LittleEndian.PutUint16(b, v)
}

// PutUint32 encodes v into b.
func (littleEndian) PutUint32(b []byte, v uint32) {
	binary.LittleEndian.PutUint32(b, v)
}

// PutUint64 encodes v into b.
func (littleEndian) PutUint64(b []byte, v uint64) {
	binary.LittleEndian.PutUint64(b, v)
}

// PutFloat encodes v into b.
func (littleEndian) PutFloat(b []byte, v float32) {
	// TODO, 当前未使用
}

// PutDouble encodes v into b.
func (littleEndian) PutDouble(b []byte, v float64) {
	// TODO, 当前未使用
}

// String returns the string representation of b.
func (littleEndian) String() string {
	return binary.LittleEndian.String()
}
