package monitor

import (
	"agent/utils/flog"
)

var (
	filterErrorLog *flog.Filter
)
// LogReportError 上报错误日志
func LogReportError(reportErr error) {
	filterErrorLog.Errorf(reportErr.Error(), "report error: %v", reportErr)
}
