package otlplog

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

var tracer = otel.Tracer("rawotlp")

func TraceSpanFields(ctx context.Context) []zap.Field {
	span := trace.SpanContextFromContext(ctx)
	if !span.IsValid() {
		fmt.Printf("no span in context %v\n", ctx)
		return nil
	}
	fmt.Printf("adding trace_id+span_id: %v\n", ctx)

	return []zap.Field{
		zap.String("trace_id", span.TraceID().String()),
		zap.String("span_id", span.SpanID().String()),
	}
}
