// Package errs provides ...
package errs

import (
	"net/http"
	thttp "trpc.group/trpc-go/trpc-go/http"
)

const (
	// ETRPCKey etrpc异常上报 key，用来做监控报警
	ETRPCKey = "etrpc_report"
)

func init() {
	// 未知错误
	thttp.RegisterStatus(ErrUnknown, http.StatusInternalServerError)
}
