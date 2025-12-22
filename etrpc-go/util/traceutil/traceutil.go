// Package traceutil provides various string tools
package traceutil

import (
	"context"

	"go.opentelemetry.io/otel/trace"
	"trpc.group/trpc-go/trpc-go"
)

// GetTraceID  获取 trace id
func GetTraceID(ctx context.Context) string {
	span := trace.SpanContextFromContext(ctx)
	if span.IsValid() {
		return span.TraceID().String()
	}
	return ""
}

// GetRequestID 获取
func GetRequestID(ctx context.Context) uint32 {
	msg := trpc.Message(ctx)
	return msg.RequestID()
}
