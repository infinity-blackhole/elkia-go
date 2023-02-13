package main

import (
	"github.com/infinity-blackhole/elkia/internal/authserver"
	fleetv1alpha1pb "github.com/infinity-blackhole/elkia/pkg/api/fleet/v1alpha1"
	"github.com/infinity-blackhole/elkia/pkg/nostale"
	"google.golang.org/grpc"

	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

func main() {
	conn, err := grpc.Dial("unix:///var/run/elkia.sock", grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	srv := nostale.NewServer(nostale.ServerConfig{
		Addr: ":8080",
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
