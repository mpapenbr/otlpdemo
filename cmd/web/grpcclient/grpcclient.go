package grpcclient

import (
	"context"
	"fmt"

	pb "buf.build/gen/go/mpapenbr/petapis/grpc/go/pet/v1/petv1grpc"
	petv1 "buf.build/gen/go/mpapenbr/petapis/protocolbuffers/go/pet/v1"
	"github.com/spf13/cobra"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	"github.com/mpapenbr/otlpdemo/cmd/config"
	"github.com/mpapenbr/otlpdemo/log"
)

func NewSimpleGRPCClientCommand() *cobra.Command {
	cmd := cobra.Command{
		Use:   "grpcclient",
		Short: "create a simple gRPC client",
		Long:  ``,
		Run: func(cmd *cobra.Command, args []string) {
			simpleGRPCClient()
		},
	}
	cmd.Flags().StringVar(&config.Address,
		"addr",
		"localhost:8080",
		"connect to this server address")

	return &cmd
}

//nolint:funlen // ok by design
func simpleGRPCClient() {
	fmt.Printf("Starting gRPC connection to %s\n", config.Address)
	myTLS, err := config.BuildTLSConfig()
	if err != nil {
		log.Error("TLS config error", log.ErrorField(err))
		return
	}
	var creds credentials.TransportCredentials
	if myTLS == nil {
		creds = insecure.NewCredentials()
	} else {
		creds = credentials.NewTLS(myTLS)
	}

	conn, err := grpc.NewClient(config.Address,
		grpc.WithTransportCredentials(creds),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	)
	if err != nil {
		log.Error("error creating connection", log.ErrorField(err))
		return
	}
	defer conn.Close()

	log.Debug("gRPC connection established", log.String("address", config.Address))
	client := pb.NewPetStoreServiceClient(conn)
	req := &petv1.GetPetRequest{
		PetId: "1234",
	}
	// not really used here, just to show how to set metadata
	sendMD := metadata.Pairs(
		"client-id", "otlpdemo-client",
	)
	ctx := metadata.NewOutgoingContext(context.Background(), sendMD)
	var recvMDHeader, recvMDTrailer metadata.MD
	resp, err := client.GetPet(ctx, req,
		grpc.Header(&recvMDHeader),
		grpc.Trailer(&recvMDTrailer))
	if err != nil {
		log.Error("error getting pet",
			log.String("trace-id", recvMDHeader.Get("trace-id")[0]),
			log.ErrorField(err))
		return
	}
	log.Debug("gRPC request done",
		log.Any("receivedHeader", recvMDHeader),
		log.Any("receivedTrailer", recvMDTrailer),
		log.String("petName", resp.Pet.Name),
	)
}
