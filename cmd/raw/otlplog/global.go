package otlplog

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var tracer = otel.Tracer("rawotlp")

// use this to control minsev LogProcesso
type minsevSeverity struct{ severity log.Severity }

func (m *minsevSeverity) Severity() log.Severity { return m.severity }

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

func convertLevel(level zapcore.Level) log.Severity {
	switch level {
	case zapcore.DebugLevel:
		return log.SeverityDebug
	case zapcore.InfoLevel:
		return log.SeverityInfo
	case zapcore.WarnLevel:
		return log.SeverityWarn
	case zapcore.ErrorLevel:
		return log.SeverityError
	case zapcore.DPanicLevel:
		return log.SeverityFatal1
	case zapcore.PanicLevel:
		return log.SeverityFatal2
	case zapcore.FatalLevel:
		return log.SeverityFatal3
	case zapcore.InvalidLevel:
		return log.SeverityUndefined
	default:
		return log.SeverityUndefined
	}
}
