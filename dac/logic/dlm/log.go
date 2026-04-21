// Package dlm 提供分布式锁管理，确保同一时刻只有一个实例运行采集任务。
package dlm

import (
	"fmt"

	"dac/entity/config"
)

// Infof 输出带Worker ID前缀的Info级别日志
func (w *Worker) Infof(format string, args ...interface{}) {
	config.Log.Infof(fmt.Sprintf("[worker %v]: ", w.id)+format, args...)
}

// Warnf 输出带Worker ID前缀的Warn级别日志
func (w *Worker) Warnf(format string, args ...interface{}) {
	config.Log.Warnf(fmt.Sprintf("[worker %v]: ", w.id)+format, args...)
}
