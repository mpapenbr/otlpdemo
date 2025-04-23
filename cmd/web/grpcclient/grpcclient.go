package grpcclient

import (
	"context"
	"fmt"

	pb "buf.build/gen/go/mpapenbr/petapis/grpc/go/pet/v1/petv1grpc"
	petv1 "buf.build/gen/go/mpapenbr/petapis/protocolbuffers/go/pet/v1"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

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
	cmd.Flags().StringVar(&config.Address, "addr", "localhost:8080", "listen address")

	return &cmd
}

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

	conn, err := grpc.NewClient(config.Address, grpc.WithTransportCredentials(creds))
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
	resp, err := client.GetPet(context.Background(), req)
	if err != nil {
		log.Error("error getting pet", log.ErrorField(err))
		return
	}
	log.Debug("gRPC request done",

		log.String("petName", resp.Pet.Name),
	)
}
