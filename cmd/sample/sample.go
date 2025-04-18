package sample

import (
	"context"
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/mpapenbr/otlpdemo/log"
	"github.com/spf13/cobra"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

var (
	duration     time.Duration
	pause        time.Duration
	work         time.Duration
	workSegments int32
)

func NewSampleCommand() *cobra.Command {
	cmd := cobra.Command{
		Use:   "sample",
		Short: "provides (random) sample telementry data",
		Long:  ``,
		RunE: func(cmd *cobra.Command, args []string) error {
			return produceSampleData()
		},
	}
	cmd.Flags().DurationVar(&duration,
		"duration",
		time.Duration(time.Minute*5),
		"duration of producing sample data")
	cmd.Flags().DurationVar(&pause,
		"pause",
		time.Duration(time.Second*15),
		"pause between samples")
	cmd.Flags().DurationVar(&work,
		"work",
		time.Duration(time.Second*3),
		"simulated work time")
	cmd.Flags().Int32Var(&workSegments,
		"segments",
		5,
		"simulated work segements")
	return &cmd
}

func produceSampleData() error {
	meter := otel.Meter("sample")
	apiCounter, err := meter.Int64Counter("sample.counter",
		metric.WithDescription("Number of calls"),
		metric.WithUnit("{call}"))
	if err != nil {
		return err
	}

	apiDurations, err := meter.Float64Histogram("sample.duration",
		metric.WithDescription("The duration of task execution"),
		metric.WithUnit("s"))
	if err != nil {
		return err
	}
	start := time.Now()
	for time.Since(start) < duration {
		doit(apiCounter, apiDurations) //nolint:errcheck //temp
		log.Debug("pausing", log.Duration("pause", pause))
		time.Sleep(pause)
	}
	return nil
}

func doit(apiCounter metric.Int64Counter, apiDuration metric.Float64Histogram) error {
	tracer := otel.Tracer("sampleData")
	start := time.Now()
	myCtx, span := tracer.Start(context.Background(), "sampleData")
	defer span.End()

	//
	segs := rand.Int32N(workSegments) + 1
	attrs := []attribute.KeyValue{
		attribute.Int("segments", int(segs)),
	}
	span.SetAttributes(attrs...)
	for i := range segs {
		workTime := rand.Int64N(int64(work))
		attrs := []attribute.KeyValue{
			attribute.Int("segment", int(i)),
			attribute.Int64("work", workTime),
		}
		_, lSpan := tracer.Start(myCtx, fmt.Sprintf("worker segmemt %d", i))
		lSpan.SetAttributes(attrs...)
		time.Sleep(time.Duration(workTime))
		log.Debug("work done",
			log.Int32("segment", i),
			log.Duration("duration", time.Duration(workTime)))
		lSpan.End()
	}

	apiCounter.Add(context.Background(), 1)
	apiDuration.Record(context.Background(), (time.Since(start)).Seconds())
	return nil
}
