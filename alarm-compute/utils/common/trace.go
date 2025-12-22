package common

import (
	"context"

	"go.opentelemetry.io/otel"
)

var tracer = otel.Tracer("")

// TracesCustomSpanDemo 自定义span示例
func TracesCustomSpanDemo(ctx context.Context) context.Context {
	ctx, span := tracer.Start(ctx, "data_query")
	defer span.End()
	return ctx
}
