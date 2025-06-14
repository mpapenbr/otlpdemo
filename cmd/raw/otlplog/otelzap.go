package otlplog

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"go.opentelemetry.io/contrib/bridges/otelzap"
	"go.opentelemetry.io/contrib/processors/minsev"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/mpapenbr/otlpdemo/cmd/config"
	"github.com/mpapenbr/otlpdemo/otel"
)

func NewOtelZapCommand() *cobra.Command {
	cmd := cobra.Command{
		Use:   "otelzap",
		Short: "simple test via combined zap and otelzap",
		Long: `
Uses a zap logger (created and configured here) in combination with otelzap.
Note: If you see messages in loki that shouldn't be there, check the setup in rootCmd
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return doOtelZapLog()
		},
	}
	return &cmd
}

//nolint:funlen // ok here
func doOtelZapLog() error {
	ctx := context.Background()
	t, err := otel.SetupTelemetry(
		otel.WithTelemetryOutput(otel.ParseTelemetryOutput(config.OtelOutput)),
		otel.WithTelemetryContext(ctx),
	)
	if err != nil {
		return fmt.Errorf("could not setup telemetry: %w", err)
	}

	cfg := zap.NewProductionConfig()
	cfg.Level, _ = zap.ParseAtomicLevel(config.LogLevel)
	logger, _ := cfg.Build()
	otelSeverity := &minsevSeverity{convertLevel(cfg.Level.Level())}
	//nolint:errcheck // no error check for example
	defer logger.Sync()
	//nolint:whitespace // editor/linter issue
	customLogger := t.CustomizedLogger(func(
		exporter sdklog.Exporter,
		downstream sdklog.Processor,
	) sdklog.LoggerProviderOption {
		proc := minsev.NewLogProcessor(downstream, otelSeverity)
		return sdklog.WithProcessor(proc)
	})
	xCore := zapcore.NewTee(
		// logger.Core().With(TraceSpanFields(ctx)),
		logger.Core(),
		otelzap.NewCore("dings", otelzap.WithLoggerProvider(customLogger)),
	)
	useLogger := zap.New(xCore)
	useLogger.Info("standard zap message")

	spanCtx, span := tracer.Start(ctx, "testspan")
	defer span.End()

	// putting the span context into the logger instructs otelzap to use the
	// span context for the log record and therefore the traceID is set
	// note: zap logger wouldn't print the message (since debug) but otelzap does emit
	useLogger.Debug("zap debug message in span",
		zap.String("someLogAttr", "someValue"),
		zap.Any("spanCtx", spanCtx),
	)
	// putting the span context into the logger instructs otelzap to use the
	// span context for the log record and therefore the traceID is set
	// note: zap logger wouldn't print the message (since debug) but otelzap does emit
	useLogger.Info("zap info message in span",
		zap.String("someLogAttr", "someValue"),
		zap.Any("spanCtx", spanCtx),
	)

	useLogger.Warn("zap warn message in span",
		zap.String("someLogAttr", "someValue"),
		zap.Any("spanCtx", spanCtx),
	)
	span.End()
	return nil
}
