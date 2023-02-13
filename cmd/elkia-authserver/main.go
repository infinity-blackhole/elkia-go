package main

import (
	"github.com/infinity-blackhole/elkia/internal/authserver"
	fleetv1alpha1pb "github.com/infinity-blackhole/elkia/pkg/api/fleet/v1alpha1"
	"github.com/infinity-blackhole/elkia/pkg/nostale"
	"github.com/spf13/pflag"
	"google.golang.org/grpc"

	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

var (
	address                string
	elkiaApiServerEndpoint string
)

func init() {
	pflag.StringVar(
		&address,
		"address",
		":4123",
		"Address",
	)
	pflag.StringVar(
		&elkiaApiServerEndpoint,
		"elkia-api-server-endpoint",
		"localhost:8080",
		"Elkia API Server endpoint",
	)
}

func main() {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	srv := nostale.NewServer(nostale.ServerConfig{
		Addr: address,
		Handler: authserver.NewHandler(authserver.HandlerConfig{
			FleetClient: fleetv1alpha1pb.NewFleetServiceClient(conn),
		}),
	})
	if err != nil {
		panic(err)
	}
	if err := srv.ListenAndServe(); err != nil {
		panic(err)
	}
}
