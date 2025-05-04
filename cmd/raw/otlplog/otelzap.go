package otlplog

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/mpapenbr/otlpdemo/cmd/config"
	rawConfig "github.com/mpapenbr/otlpdemo/cmd/raw/config"
	"github.com/mpapenbr/otlpdemo/log"
)

func NewLogOtelZapCommand() *cobra.Command {
	cmd := cobra.Command{
		Use:   "otelzap",
		Short: "simple test via otelzap",
		Long:  ``,
		RunE: func(cmd *cobra.Command, args []string) error {
			return doOtelZapLog()
		},
	}
	return &cmd
}

func doOtelZapLog() error {
	ctx := context.Background()
	t, err := config.SetupTelemetry(
		config.WithTelemetryOutput(config.ParseTelemetryOutput(rawConfig.OutputArg)),
		config.WithTelemetryContext(ctx),
	)
	if err != nil {
		return fmt.Errorf("could not setup telemetry: %w", err)
	}
	logger, _ := log.NewOtelCore(t.LoggerProvider())

	logger.Info("standard zap message without context")

	spanCtx, span := tracer.Start(ctx, "testspan otelzap")
	defer span.End()

	logger.InfoContext(spanCtx, "otelzap message in span with context")
	span.End()
	t.Shutdown()
	return nil
}
