package main

import (
	"fmt"
	"net"
	"os"

	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.shikanime.studio/elkia/internal/auth"
	"go.shikanime.studio/elkia/internal/clients"
	eventing "go.shikanime.studio/elkia/pkg/api/eventing/v1alpha1"
	"google.golang.org/grpc"

	_ "go.shikanime.studio/elkia/internal/monitoring"
)

func main() {
	fleetCs, err := clients.NewClientSet()
	if err != nil {
		logrus.Fatal(err)
	}
	srv := grpc.NewServer(
		grpc.UnaryInterceptor(otelgrpc.UnaryServerInterceptor()),
		grpc.StreamInterceptor(otelgrpc.StreamServerInterceptor()),
	)
	host := os.Getenv("HOST")
	if host == "" {
		host = "localhost"
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%s", host, port))
	if err != nil {
		logrus.Fatal(err)
	}
	eventing.RegisterAuthServer(
		srv,
		auth.NewServer(auth.ServerConfig{
			PresenceClient: fleetCs.PresenceClient,
			ClusterClient:  fleetCs.ClusterClient,
		}),
	)
	logrus.Debugf("auth server: listening on %s:%s", host, port)
	if err := srv.Serve(lis); err != nil {
		logrus.Fatal(err)
	}
}
