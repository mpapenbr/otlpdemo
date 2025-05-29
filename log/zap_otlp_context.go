package log

import (
	"context"
	"os"
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

// This is a simple wrapper for zap which also emits the logs to the OTLP logger
// It is not a full implementation of the zap interface, but it is enough for
// our use case
// In order to link logs to traces the user must provide the span context by calling
// the InfoContext method
func NewZapWithContextBasedOTLP(provider otellog.LoggerProvider) (*ZapWithOTLP, error) {
	cfg := zap.NewProductionConfig()
	cfg.EncoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format("2006-01-02T15:04:05.000Z0700"))
	}

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(cfg.EncoderConfig),
		zapcore.AddSync(os.Stdout),
		zapcore.InfoLevel,
	)

	logger := zap.New(core, zap.AddCallerSkip(1))

	ret := &ZapWithOTLP{
		l:          logger,
		level:      zapcore.InfoLevel,
		otlpLogger: provider.Logger("contextbasedzap"),
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

// InfoContext is a wrapper for the zap logger which also emits the logs
// to the OTLP logger
// Note: here we don't convert the fields to OTLP attributes
func (c *ZapWithOTLP) InfoContext(ctx context.Context, msg string, fields ...Field) {
	if c.level.Enabled(zap.InfoLevel) {
		c.l.Info(msg, fields...)
		c.otlpLogger.Emit(ctx, c.createRecord(msg))
	}
}
