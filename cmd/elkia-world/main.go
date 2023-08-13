package main

import (
	"fmt"
	"net"
	"os"

	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.shikanime.studio/elkia/internal/clients"
	"go.shikanime.studio/elkia/internal/lobby"
	eventing "go.shikanime.studio/elkia/pkg/api/eventing/v1alpha1"
	"google.golang.org/grpc"
	"gorm.io/gorm"

	_ "go.shikanime.studio/elkia/internal/monitoring"
)

func main() {
	logrus.Debugf("Starting world server")
	dialector, err := clients.NewGormDialector()
	if err != nil {
		logrus.Fatal(err)
	}
	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		logrus.Fatal(err)
	}
	logrus.Debugf("worldserver: connected to db")
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
	eventing.RegisterLobbyServer(
		srv,
		lobby.NewLobbyServer(lobby.LobbyServerConfig{
			DB: db,
		}),
	)
	logrus.Debugf("auth server: listening on %s:%s", host, port)
	if err := srv.Serve(lis); err != nil {
		logrus.Fatal(err)
	}
}
