package byteorder

import (
	"encoding/binary"
	"math"
)

var BigEndianExtend bigEndian

type bigEndian struct{}

// Uint16 returns the uint16 encoded in b.
func (bigEndian) Uint16(b []byte) uint16 {
	return binary.BigEndian.Uint16(b)
}

// Uint32 returns the uint32 encoded in b.
func (bigEndian) Uint32(b []byte) uint32 {
	return binary.BigEndian.Uint32(b)
}

// Uint64 returns the uint64 encoded in b.
func (bigEndian) Uint64(b []byte) uint64 {
	return binary.BigEndian.Uint64(b)
}

// Float returns the float32 encoded in b.
func (be bigEndian) Float(b []byte) float32 {
	return math.Float32frombits(be.Uint32(b))
}

// Double returns the float64 encoded in b.
func (be bigEndian) Double(b []byte) float64 {
	return math.Float64frombits(be.Uint64(b))
}

// PutUint16 encodes v into b.
func (bigEndian) PutUint16(b []byte, v uint16) {
	binary.BigEndian.PutUint16(b, v)
}

// PutUint32 encodes v into b.
func (bigEndian) PutUint32(b []byte, v uint32) {
	binary.BigEndian.PutUint32(b, v)
}

// PutUint64 encodes v into b.
func (bigEndian) PutUint64(b []byte, v uint64) {
	binary.BigEndian.PutUint64(b, v)
}

// PutFloat encodes v into b.
func (bigEndian) PutFloat(b []byte, v float32) {
	// TODO, 当前未使用
}

// PutDouble encodes v into b.
func (bigEndian) PutDouble(b []byte, v float64) {
	// TODO, 当前未使用
}

// String returns the string representation of b.
func (bigEndian) String() string {
	return binary.BigEndian.String()
}
