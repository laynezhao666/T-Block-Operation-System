package utils

import (
	"sync/atomic"
)

var (
	seq uint64 = 0
)

// GetNextSequenceNumber 获取下一个序列号
func GetNextSequenceNumber() uint64 {
	return atomic.AddUint64(&seq, 1) - 1
}
