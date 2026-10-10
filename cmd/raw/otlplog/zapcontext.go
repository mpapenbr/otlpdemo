package otlplog

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"go.opentelemetry.io/otel"

	"github.com/mpapenbr/otlpdemo/cmd/config"
	"github.com/mpapenbr/otlpdemo/log"
	myOtel "github.com/mpapenbr/otlpdemo/otel"
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
	t, err := myOtel.SetupTelemetry(
		myOtel.WithTelemetryOutput(myOtel.ParseTelemetryOutput(config.OtelOutput)),
		myOtel.WithTelemetryContext(ctx),
	)
	if err != nil {
		return fmt.Errorf("could not setup telemetry: %w", err)
	}
	logger, _ := log.NewZapWithContextBasedOTLP(otel.GetLoggerProvider())

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
