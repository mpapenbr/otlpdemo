package otlplog

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"go.opentelemetry.io/otel/log/global"

	"github.com/mpapenbr/otlpdemo/cmd/config"
	"github.com/mpapenbr/otlpdemo/log"
	"github.com/mpapenbr/otlpdemo/otel"
)

func NewZapContextCommand() *cobra.Command {
	cmd := cobra.Command{
		Use:   "zapcontext",
		Short: "simple test via context base zap",
		Long:  ``,
		RunE: func(cmd *cobra.Command, args []string) error {
			return doZapContextLog()
		},
	}
	return &cmd
}

func doZapContextLog() error {
	ctx := context.Background()
	t, err := otel.SetupTelemetry(
		otel.WithTelemetryOutput(otel.ParseTelemetryOutput(config.OtelOuput)),
		otel.WithTelemetryContext(ctx),
	)
	if err != nil {
		return fmt.Errorf("could not setup telemetry: %w", err)
	}
	logger, _ := log.NewZapWithContextBasedOTLP(global.GetLoggerProvider())

	logger.Info("standard zapcontext message without context")

	spanCtx, span := tracer.Start(ctx, "testspan zapcontext")
	defer span.End()

	// you wouldn't see the attributes in OTLP since they are not handled
	logger.InfoContext(spanCtx, "zapcontext message in span with context",
		log.String("someLogAttr", "someValue"))
	span.End()
	t.Shutdown()
	return nil
}
