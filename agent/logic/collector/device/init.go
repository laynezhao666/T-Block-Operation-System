package device

import (
	"agent/utils/flog"
	"time"

	"trpc.group/trpc-go/trpc-go/log"
)

// Init 初始化
func Init() {
	filterLog = flog.NewFilterLogger(time.Minute, log.GetDefaultLogger())
}
