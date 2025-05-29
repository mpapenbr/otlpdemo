package otlplog

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/mpapenbr/otlpdemo/log"
)

func NewOwnLoggerCommand() *cobra.Command {
	cmd := cobra.Command{
		Use:   "own",
		Short: "simple test via own logger",
		Long:  ``,
		RunE: func(cmd *cobra.Command, args []string) error {
			return doOwnLogger(cmd.Context())
		},
	}
	return &cmd
}

func doOwnLogger(ctx context.Context) error {
	rootLogger := log.GetFromContext(ctx)
	if rootLogger == nil {
		return errors.New("no logger in context")
	}
	rootLogger.Info("standard own message")

	spanCtx, span := tracer.Start(ctx, "testspan root logger")

	// putting the span context into the logger instructs otelzap to use the
	// span context for the log record and therefore the traceID is set
	// note: zap logger wouldn't print the message (since debug) but otelzap does emit
	doTheLog(spanCtx, rootLogger, "root")
	span.End()

	spanCtx, span = tracer.Start(ctx, "testspan on demoLogger")
	demoLogger := rootLogger.Named("demoLogger")
	doTheLog(spanCtx, demoLogger, "demoLogger")
	span.End()

	spanCtx, span = tracer.Start(ctx, "testspan on otherLogger")
	otherLogger := rootLogger.Named("otherLogger")
	doTheLog(spanCtx, otherLogger, "otherLogger")
	span.End()
	//nolint:errcheck // no error check for example
	log.Sync()
	return nil
}

func doTheLog(spanCtx context.Context, useLogger *log.Logger, extra string) {
	useLogger.Debug(
		fmt.Sprintf("standard own DEBUG message in span (%s)", extra),
		zap.String("someLogAttr", "someValue"),
		zap.Any("spanCtx", spanCtx),
	)
	useLogger.Info(
		fmt.Sprintf("standard own INFO message in span (%s)", extra),
		zap.String("someLogAttr", "someValue"),
		zap.Any("spanCtx", spanCtx),
	)
	useLogger.Warn(
		fmt.Sprintf("standard own WARN message in span (%s)", extra),
		zap.String("someLogAttr", "someValue"),
		zap.Any("spanCtx", spanCtx),
	)
	useLogger.Error(
		fmt.Sprintf("standard own ERROR message in span (%s)", extra),
		zap.String("someLogAttr", "someValue"),
		zap.Any("spanCtx", spanCtx),
	)
}
