package otlplog

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"go.opentelemetry.io/contrib/bridges/otelzap"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/mpapenbr/otlpdemo/cmd/config"
	rawConfig "github.com/mpapenbr/otlpdemo/cmd/raw/config"
)

func NewLogZapCommand() *cobra.Command {
	cmd := cobra.Command{
		Use:   "zap",
		Short: "simple test via combined zap and otelzap",
		Long:  ``,
		RunE: func(cmd *cobra.Command, args []string) error {
			return doZapLog()
		},
	}
	return &cmd
}

func doZapLog() error {
	ctx := context.Background()
	t, err := config.SetupTelemetry(
		config.WithTelemetryOutput(config.ParseTelemetryOutput(rawConfig.OutputArg)),
		config.WithTelemetryContext(ctx),
	)
	if err != nil {
		return fmt.Errorf("could not setup telemetry: %w", err)
	}
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	xCore := zapcore.NewTee(
		logger.Core().With(TraceSpanFields(ctx)),
		otelzap.NewCore("dings", otelzap.WithLoggerProvider(t.LoggerProvider())),
	)
	useLogger := zap.New(xCore)
	useLogger.Info("standard zap message")

	spanCtx, span := tracer.Start(ctx, "testspan")
	defer span.End()

	fmt.Printf("ctx: %v\n", ctx)
	fmt.Printf("spanCtx: %v\n", spanCtx)

	useLogger.Info("standard zap message in span")
	span.End()
	t.Shutdown()
	return nil
}
