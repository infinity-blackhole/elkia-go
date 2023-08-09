package main

import (
	"fmt"
	"net"
	"net/url"
	"os"

	"github.com/infinity-blackhole/elkia/internal/lobby"
	world "github.com/infinity-blackhole/elkia/pkg/api/world/v1alpha1"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	_ "github.com/infinity-blackhole/elkia/internal/monitoring"
)

func main() {
	logrus.Debugf("Starting world server")
	dsnUrlStr := os.Getenv("DSN")
	if dsnUrlStr == "" {
		logrus.Fatal("worldserver: DSN is not set")
	}
	dsnUrl, err := url.Parse(dsnUrlStr)
	if err != nil {
		logrus.Fatal(err)
	}
	db, err := gorm.Open(mysql.Open(dsnUrl.String()), &gorm.Config{})
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
	world.RegisterLobbyServer(
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
