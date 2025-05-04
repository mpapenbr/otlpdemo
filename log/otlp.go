package log

import (
	"context"
	"time"

	otellog "go.opentelemetry.io/otel/log"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type ZapWithOTLP struct {
	l          *zap.Logger
	level      zapcore.Level
	otlpLogger otellog.Logger
}

func NewOtelCore(provider otellog.LoggerProvider) (*ZapWithOTLP, error) {
	l, err := zap.NewProduction()
	if err != nil {
		return nil, err
	}
	ret := &ZapWithOTLP{
		l:          l,
		level:      zapcore.DebugLevel,
		otlpLogger: provider.Logger("otelzap"),
	}

	return ret, nil
}

func (c *ZapWithOTLP) Info(msg string, fields ...Field) {
	c.InfoContext(context.Background(), msg, fields...)
}

func (c *ZapWithOTLP) createRecord(msg string) otellog.Record {
	var r otellog.Record
	r.SetTimestamp(time.Now())
	r.SetSeverity(otellog.SeverityInfo)
	r.SetBody(otellog.StringValue(msg))
	return r
}

func (c *ZapWithOTLP) InfoContext(ctx context.Context, msg string, fields ...Field) {
	c.l.Info(msg, fields...)
	c.otlpLogger.Emit(ctx, c.createRecord(msg))
}
