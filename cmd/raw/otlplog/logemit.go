package otlplog

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/log/global"

	"github.com/mpapenbr/otlpdemo/cmd/config"
	"github.com/mpapenbr/otlpdemo/otel"
)

func NewLogEmitCommand() *cobra.Command {
	cmd := cobra.Command{
		Use:   "emit",
		Short: "simple test via direct emit with record",
		Long:  ``,
		RunE: func(cmd *cobra.Command, args []string) error {
			return doLogEmit()
		},
	}
	return &cmd
}

func doLogEmit() error {
	ctx := context.Background()
	t, err := otel.SetupTelemetry(
		otel.WithTelemetryOutput(otel.ParseTelemetryOutput(config.OtelOuput)),
		otel.WithTelemetryContext(ctx),
	)
	if err != nil {
		return fmt.Errorf("could not setup telemetry: %w", err)
	}

	fmt.Println("doLogtest")
	logger := global.GetLoggerProvider().Logger("testlogger")

	// Emit a log
	r := createRecord("test message")
	logger.Emit(ctx, r)

	spanCtx, span := tracer.Start(ctx, "testspan")
	defer span.End()
	span.AddEvent("creating log record in span")
	r = createRecord("test message in span")
	// In order to get the correct traceID we must use the span context here
	// otherwise the traceID is empty and it can't be linked to the trace
	logger.Emit(spanCtx, r)

	span.End()
	t.Shutdown()
	return nil
}

func createRecord(msg string) log.Record {
	var r log.Record
	r.SetTimestamp(time.Now())
	r.SetSeverity(log.SeverityInfo)
	r.SetBody(log.StringValue(msg))
	return r
}
