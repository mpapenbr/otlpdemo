//nolint:gosec // ignore G404 (rand) mainly
package grpcserver

import (
	"context"
	"fmt"
	"math/rand/v2"
	"net"
	"time"

	pb "buf.build/gen/go/mpapenbr/petapis/grpc/go/pet/v1/petv1grpc"
	petv1 "buf.build/gen/go/mpapenbr/petapis/protocolbuffers/go/pet/v1"
	"github.com/spf13/cobra"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/mpapenbr/otlpdemo/cmd/config"
	"github.com/mpapenbr/otlpdemo/log"
)

var tracer = otel.Tracer("grcpserver")

func NewSimpleGRPCServerCommand() *cobra.Command {
	cmd := cobra.Command{
		Use:   "grpcserver",
		Short: "create a simple gRPC server",
		Long:  ``,
		Run: func(cmd *cobra.Command, args []string) {
			simpleGRPCserver()
		},
	}
	cmd.Flags().StringVar(&config.Address, "addr", ":8080", "listen address")

	return &cmd
}

func simpleGRPCserver() {
	fmt.Printf("Starting server on %s\n", config.Address)
	creds, err := config.BuildTransportCredentials()
	if err != nil {
		log.Error("TLS config error", log.ErrorField(err))
		return
	}

	var l net.ListenConfig
	lis, err := l.Listen(context.Background(), "tcp", config.Address)
	if err != nil {
		log.Error("error starting listener", log.ErrorField(err))
		return
	}
	statsHandler := grpc.StatsHandler(otelgrpc.NewServerHandler())

	srv := grpc.NewServer(grpc.Creds(creds), statsHandler,
		grpc.ChainUnaryInterceptor(
			TraceIDHeaderInterceptor(),
		))
	pb.RegisterPetStoreServiceServer(srv, &petServer{})
	if err := srv.Serve(lis); err != nil {
		log.Error("error starting server", log.ErrorField(err))
		return
	}
}

// TraceIDHeaderInterceptor injects the trace ID into gRPC response headers.
//
//nolint:whitespace // editor/linter issue
func TraceIDHeaderInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		span := trace.SpanFromContext(ctx)
		if span != nil && span.SpanContext().IsValid() {
			traceID := span.SpanContext().TraceID().String()
			_ = grpc.SetHeader(ctx, metadata.Pairs("trace-id", traceID))
		}

		return handler(ctx, req)
	}
}

type petServer struct {
	pb.UnimplementedPetStoreServiceServer
}

var _ pb.PetStoreServiceServer = (*petServer)(nil)

//nolint:whitespace // editor/linter issue
func (s *petServer) GetPet(ctx context.Context, req *petv1.GetPetRequest) (
	*petv1.GetPetResponse, error,
) {
	log.Debug("GetPet called", log.String("petId", req.PetId), log.Any("spanCtx", ctx))
	span := trace.SpanFromContext(ctx)
	s.ringTheBell(ctx)
	if pet, err := s.lookingForRequestedPet(ctx, req.PetId); err != nil {
		log.Error("pet not found", log.String("petId", req.PetId), log.ErrorField(err))
		span.SetStatus(codes.Error, "pet could not be found")
		return nil, err
	} else {
		span.SetStatus(codes.Ok, "pet found and returned")
		return &petv1.GetPetResponse{Pet: pet}, nil
	}
}

func (s *petServer) ringTheBell(ctx context.Context) {
	// This is just a dummy function to show how to use the tracer
	// in a real application, you would do something useful here
	spanCtx, span := tracer.Start(ctx, "ringing the bell")
	defer span.End()
	log.Debug("ringTheBell called", log.Any("spanCtx", spanCtx))
	time.Sleep(20 * time.Millisecond) // Simulate some work
	span.AddEvent("bell found")
	time.Sleep(100 * time.Millisecond) // Simulate some work
	log.Debug("clerk arrived ", log.Any("spanCtx", spanCtx))
}

//nolint:whitespace // editor/linter issue
func (s *petServer) lookingForRequestedPet(ctx context.Context, petID string) (
	*petv1.Pet, error,
) {
	// This is just a dummy function to show how to use the tracer
	// in a real application, you would do something useful here
	spanCtx, span := tracer.Start(ctx, "looking for requested pet",
		trace.WithAttributes(
			attribute.String("petId", petID),
		))
	defer span.End()
	span.SetAttributes()
	log.Debug("lookgingForRequestedPet called", log.Any("spanCtx", spanCtx))
	time.Sleep(50 * time.Millisecond) // Simulate some work
	if rand.IntN(10) == 0 {
		span.AddEvent("pet not found")
		span.SetStatus(codes.Error, "pet not found")
		return nil, fmt.Errorf("pet with ID %s not found", petID)
	}
	return &petv1.Pet{
		PetId:   petID,
		Name:    "Fluffy",
		PetType: petv1.PetType_PET_TYPE_CAT,
	}, nil
}
