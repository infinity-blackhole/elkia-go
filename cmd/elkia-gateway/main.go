package main

import (
	"fmt"
	"net"
	"os"

	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.shikanime.studio/elkia/internal/clients"
	"go.shikanime.studio/elkia/internal/gateway"
	eventingpb "go.shikanime.studio/elkia/pkg/api/eventing/v1alpha1"
	"google.golang.org/grpc"
)

func init() {
	if logLevelStr := os.Getenv("LOG_LEVEL"); logLevelStr != "" {
		logLevel, err := logrus.ParseLevel(logLevelStr)
		if err != nil {
			logrus.Fatal(err)
		}
		logrus.SetLevel(logLevel)
	}
}

func main() {
	fleetCs, err := clients.NewFleetClientSet()
	if err != nil {
		logrus.Fatal(err)
	}
	redisClient := clients.NewRedisClient()
	logrus.Debugf("gateway: connected to redis")
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
	eventingpb.RegisterGatewayServer(
		srv,
		gateway.NewServer(gateway.ServerConfig{
			PresenceClient: fleetCs.PresenceClient,
			RedisClient:    redisClient,
		}),
	)
	logrus.Debugf("auth server: listening on %s:%s", host, port)
	if err := srv.Serve(lis); err != nil {
		logrus.Fatal(err)
	}
}
