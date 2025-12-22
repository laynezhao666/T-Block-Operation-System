package utils

import (
	"strconv"
	"sync/atomic"
)
// MessageIDType 消息id类型
type MessageIDType = uint64

var (
	messageID MessageIDType
)

func init() {
	atomic.StoreUint64(&messageID, 0)
}
// NextMessageID 生成消息id
func NextMessageID() string {
	return strconv.FormatUint(atomic.AddUint64(&messageID, 1), 10)
}
