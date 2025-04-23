package grpcserver

import (
	"context"
	"fmt"
	"net"

	pb "buf.build/gen/go/mpapenbr/petapis/grpc/go/pet/v1/petv1grpc"
	petv1 "buf.build/gen/go/mpapenbr/petapis/protocolbuffers/go/pet/v1"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"

	"github.com/mpapenbr/otlpdemo/cmd/config"
	"github.com/mpapenbr/otlpdemo/log"
)

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

	lis, err := net.Listen("tcp", config.Address)
	if err != nil {
		log.Error("error starting listener", log.ErrorField(err))
		return
	}
	srv := grpc.NewServer(grpc.Creds(creds))
	pb.RegisterPetStoreServiceServer(srv, &petServer{})
	if err := srv.Serve(lis); err != nil {
		log.Error("error starting server", log.ErrorField(err))
		return
	}
}

type petServer struct {
	pb.UnimplementedPetStoreServiceServer
}

//nolint:whitespace // editor/linter issue
func (s *petServer) GetPet(ctx context.Context, req *petv1.GetPetRequest) (
	*petv1.GetPetResponse, error,
) {
	log.Debug("GetPet called", log.String("petId", req.PetId))
	return &petv1.GetPetResponse{
		Pet: &petv1.Pet{
			PetId:   req.PetId,
			Name:    "Fluffy",
			PetType: petv1.PetType_PET_TYPE_CAT,
		},
	}, nil
}
