package config

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdkresource "go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	"github.com/mpapenbr/otlpdemo/log"
)

var (
	resource          *sdkresource.Resource
	initResourcesOnce sync.Once
)

func SetupStdOutMetrics() (sdkmetric.Exporter, error) {
	return stdoutmetric.New()
}

func SetupStdOutTracing() (sdktrace.SpanExporter, error) {
	return stdouttrace.New()
}

type Telemetry struct {
	ctx     context.Context
	metrics *sdkmetric.MeterProvider
	traces  *sdktrace.TracerProvider
}

func (t Telemetry) Shutdown() {
	log.Info("Shutdown telemetry")
	if err := t.metrics.Shutdown(context.Background()); err != nil {
		fmt.Printf("shutdown metrics error:%+v\n", err)
	}
	if err := t.traces.Shutdown(context.Background()); err != nil {
		fmt.Printf("shutdown traces error:%+v\n", err)
	}
}

func SetupTelemetry(ctx context.Context) (*Telemetry, error) {
	ret := Telemetry{ctx: ctx}

	if m, err := setupMetrics(); err != nil {
		return nil, err
	} else {
		ret.metrics = m
	}
	if t, err := setupTraces(); err != nil {
		return nil, err
	} else {
		ret.traces = t
	}
	return &ret, nil
}

func setupMetrics() (*sdkmetric.MeterProvider, error) {
	exporter, err := otlpmetricgrpc.New(
		context.Background(),
	)
	if err != nil {
		return nil, err
	}
	provider := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(initResource()),
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(exporter,
			sdkmetric.WithInterval(15*time.Second))), // TODO: configure?
	)

	otel.SetMeterProvider(provider)
	return provider, nil
}

func setupTraces() (*sdktrace.TracerProvider, error) {
	exporter, err := otlptracegrpc.New(
		context.Background(),
	)
	if err != nil {
		return nil, err
	}
	provider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(initResource()),
		// set the sampling rate based on the parent span to 60%
		// sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(0.6))),
		// sdktrace.WithSampler(sdktrace.AlwaysSample()), // TODO: confiure?
	)

	otel.SetTracerProvider(provider)

	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			// W3C Trace Context format; https://www.w3.org/TR/trace-context/
			propagation.TraceContext{}, propagation.Baggage{},
		),
	)
	return provider, nil
}

func initResource() *sdkresource.Resource {
	initResourcesOnce.Do(func() {
		extraResources, _ := sdkresource.New(
			context.Background(),
			sdkresource.WithOS(),
			sdkresource.WithProcess(),
			sdkresource.WithContainer(),
			sdkresource.WithHost(),
		)
		resource, _ = sdkresource.Merge(
			sdkresource.Default(),
			extraResources,
		)
	})
	return resource
}
