package otlplog

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"go.opentelemetry.io/otel/log"

	"github.com/mpapenbr/otlpdemo/cmd/config"
	rawConfig "github.com/mpapenbr/otlpdemo/cmd/raw/config"
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
	t, err := config.SetupTelemetry(
		config.WithTelemetryOutput(config.ParseTelemetryOutput(rawConfig.OutputArg)),
		config.WithTelemetryContext(ctx),
	)
	if err != nil {
		return fmt.Errorf("could not setup telemetry: %w", err)
	}

	fmt.Println("doLogtest")
	logger := t.LoggerProvider().Logger("testlogger")

	// Emit a log
	r := createRecord("test message")
	logger.Emit(ctx, r)

	spanCtx, span := tracer.Start(ctx, "testspan")
	defer span.End()
	span.AddEvent("creating log record in span")
	r = createRecord("test message in span")
	logger.Emit(spanCtx, r)

	// r = createRecord("crafted with TraceID on stdCtx")
	// r.AddAttributes(log.KeyValue{
	// 	Key: "TraceID", Value: log.StringValue(span.SpanContext().TraceID().String()),
	// })
	// logger.Emit(ctx, r)
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
