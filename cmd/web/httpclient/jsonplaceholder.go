package httpclient

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/spf13/cobra"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

	"github.com/mpapenbr/otlpdemo/log"
)

func NewJSONPlaceholderCommand() *cobra.Command {
	ret := cobra.Command{
		Use:   "jsonplaceholder",
		Short: "issue requests to jsonplaceholder",
		Long:  ``,
		RunE: func(cmd *cobra.Command, args []string) error {
			return queryJSONPlaceholder()
		},
	}
	return &ret
}

func queryJSONPlaceholder() error {
	meter := otel.Meter("jsonplaceholder")
	apiCounter, err := meter.Int64Counter("api.counter",
		metric.WithDescription("Number of calls"),
		metric.WithUnit("{call}"))
	if err != nil {
		return err
	}

	apiDurations, err := meter.Float64Histogram("api.duration",
		metric.WithDescription("The duration of task execution"),
		metric.WithUnit("s"))
	if err != nil {
		return err
	}

	for i := 0; i < 30; i++ {
		doit(apiCounter, apiDurations) //nolint:errcheck //temp
		time.Sleep(1 * time.Second)
	}

	return nil
}

func doit(apiCounter metric.Int64Counter, apiDuration metric.Float64Histogram) error {
	tracer := otel.Tracer("jsonplaceholder")
	start := time.Now()
	ctx, span := tracer.Start(context.Background(), "jsonplaceholder")
	defer span.End()

	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		"https://jsonplaceholder.typicode.com/todos/1",
		http.NoBody)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	span.AddEvent("got result")
	body, err := io.ReadAll(resp.Body)
	attrs := []attribute.KeyValue{
		attribute.Int("status", resp.StatusCode),
		attribute.Int("bytes", len(body)),
	}
	span.SetAttributes(attrs...)

	if err != nil {
		return err
	}
	log.Debug("request done",
		log.Int("status", resp.StatusCode), log.Int("bytes", len(body)))
	resp.Body.Close()
	apiCounter.Add(ctx, 1)
	apiDuration.Record(ctx, (time.Since(start)).Seconds())
	return nil
}
