// Package log provides ...
// @author: xincili
// -------------------------------------------
package log

import (
	"context"
	"trpc.group/trpc-go/trpc-go/codec"
	"trpc.group/trpc-go/trpc-go/log"
)

// GetLogger 日志方法
var GetLogger = func(ctx context.Context) log.Logger {
	switch l := codec.Message(ctx).Logger().(type) {
	case log.Logger:
		return l.With(getCommonFields()...)
	default:
		return log.GetDefaultLogger()
	}
}
