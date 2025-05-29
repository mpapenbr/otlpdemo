package otel

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutlog"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/propagation"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdkresource "go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"

	"github.com/mpapenbr/otlpdemo/version"
)

var (
	resource          *sdkresource.Resource
	initResourcesOnce sync.Once
)

type (
	logConfig struct {
		exporter   sdklog.Exporter
		downstream sdklog.Processor
	}
	config struct {
		ctx       context.Context
		output    TelemetryOutput
		logConfig *logConfig
	}
	Telemetry struct {
		config  *config
		metrics *sdkmetric.MeterProvider
		traces  *sdktrace.TracerProvider
		logs    *sdklog.LoggerProvider
	}
	TelemetryOutput     int
	TelemetryOption     func(cfg *config)
	CustomizeLoggerFunc func(
		exporter sdklog.Exporter,
		downstream sdklog.Processor) sdklog.LoggerProviderOption
	CustomizedLoggerOption interface {
		apply(CustomizeLoggerFunc)
	}
)

const (
	StdOut TelemetryOutput = iota
	Grpc
)

func (to TelemetryOutput) String() string {
	switch to {
	case StdOut:
		return "stdout"
	case Grpc:
		return "grpc"
	default:
		return "unknown"
	}
}

func ParseTelemetryOutput(arg string) TelemetryOutput {
	switch strings.ToLower(arg) {
	case "stdout":
		return StdOut
	case "grpc":
		return Grpc
	default:
		return StdOut
	}
}

func WithTelemetryContext(arg context.Context) TelemetryOption {
	return func(cfg *config) {
		cfg.ctx = arg
	}
}

func WithTelemetryOutput(arg TelemetryOutput) TelemetryOption {
	return func(cfg *config) {
		cfg.output = arg
	}
}

func SetupTelemetry(opts ...TelemetryOption) (*Telemetry, error) {
	cfg := config{
		ctx:    context.Background(),
		output: Grpc,
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	ret := Telemetry{config: &cfg}

	if err := ret.setupMetrics(); err != nil {
		return nil, err
	}
	if err := ret.setupTraces(); err != nil {
		return nil, err
	}
	if err := ret.setupLogs(); err != nil {
		return nil, err
	}

	return &ret, nil
}

func (t Telemetry) Shutdown() {
	if err := t.metrics.ForceFlush(context.Background()); err != nil {
		fmt.Printf("flushing metrics error:%+v\n", err)
	}
	if err := t.traces.ForceFlush(context.Background()); err != nil {
		fmt.Printf("flushing traces error:%+v\n", err)
	}
	if err := t.logs.ForceFlush(context.Background()); err != nil {
		fmt.Printf("flushing logs error:%+v\n", err)
	}
	if err := t.metrics.Shutdown(context.Background()); err != nil {
		fmt.Printf("shutdown metrics error:%+v\n", err)
	}
	if err := t.traces.Shutdown(context.Background()); err != nil {
		fmt.Printf("shutdown traces error:%+v\n", err)
	}
	if err := t.logs.Shutdown(context.Background()); err != nil {
		fmt.Printf("shutdown logs error:%+v\n", err)
	}
}

//nolint:lll // readabilty
func (t Telemetry) CustomizedLogger(opts ...CustomizeLoggerFunc) *sdklog.LoggerProvider {
	lgOpts := make([]sdklog.LoggerProviderOption, len(opts))
	for i, opt := range opts {
		lgOpts[i] = opt(t.config.logConfig.exporter, t.config.logConfig.downstream)
	}
	return sdklog.NewLoggerProvider(lgOpts...)
}

func (t *Telemetry) setupMetrics() (err error) {
	var exporter sdkmetric.Exporter
	switch t.config.output {
	case StdOut:
		exporter, err = stdoutmetric.New()
	case Grpc:
		exporter, err = otlpmetricgrpc.New(t.config.ctx)
	}
	if err != nil {
		return err
	}
	provider := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(initResource()),
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(exporter,
			sdkmetric.WithInterval(15*time.Second))), // TODO: configure?
	)

	otel.SetMeterProvider(provider)
	t.metrics = provider
	return nil
}

func (t *Telemetry) setupTraces() (err error) {
	var exporter sdktrace.SpanExporter
	switch t.config.output {
	case StdOut:
		exporter, err = stdouttrace.New()
	case Grpc:
		exporter, err = otlptracegrpc.New(t.config.ctx)
	}
	if err != nil {
		return err
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
	t.traces = provider
	return nil
}

func (t *Telemetry) setupLogs() (err error) {
	var exporter sdklog.Exporter
	switch t.config.output {
	case StdOut:
		exporter, err = stdoutlog.New()
	case Grpc:
		exporter, err = otlploggrpc.New(t.config.ctx)
	}
	if err != nil {
		return err
	}

	proc := sdklog.NewSimpleProcessor(exporter)
	// use this if experimental phase is finished
	// proc := sdklog.NewBatchProcessor(exporter)
	t.config.logConfig = &logConfig{
		exporter:   exporter,
		downstream: proc,
	}

	provider := sdklog.NewLoggerProvider(
		sdklog.WithProcessor(proc),
	)
	global.SetLoggerProvider(provider)
	t.logs = provider
	return nil
}

func initResource() *sdkresource.Resource {
	initResourcesOnce.Do(func() {
		extraResources, _ := sdkresource.New(
			context.Background(),
			sdkresource.WithOS(),
			sdkresource.WithProcess(),
			sdkresource.WithContainer(),
			sdkresource.WithHost(),
			sdkresource.WithAttributes(semconv.ServiceVersion(version.Version)),
		)
		resource, _ = sdkresource.Merge(
			sdkresource.Default(),
			extraResources,
		)
	})
	return resource
}
