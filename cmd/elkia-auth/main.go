package main

import (
	"fmt"
	"net"
	"os"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_logrus "github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"github.com/infinity-blackhole/elkia/internal/auth"
	eventing "github.com/infinity-blackhole/elkia/pkg/api/eventing/v1alpha1"
	fleet "github.com/infinity-blackhole/elkia/pkg/api/fleet/v1alpha1"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	_ "k8s.io/client-go/plugin/pkg/client/auth"
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
	elkiaFleetEndpoint := os.Getenv("ELKIA_FLEET_ENDPOINT")
	if elkiaFleetEndpoint == "" {
		elkiaFleetEndpoint = "localhost:8080"
	}
	logrus.Debugf("auth: connecting to fleet at %s", elkiaFleetEndpoint)
	conn, err := grpc.Dial(
		elkiaFleetEndpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
		grpc.WithStreamInterceptor(otelgrpc.StreamClientInterceptor()),
	)
	if err != nil {
		logrus.Fatal(err)
	}
	srv := grpc.NewServer(
		grpc.UnaryInterceptor(
			grpc_middleware.ChainUnaryServer(
				otelgrpc.UnaryServerInterceptor(),
				grpc_ctxtags.UnaryServerInterceptor(
					grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor),
				),
				grpc_logrus.UnaryServerInterceptor(logrus.NewEntry(logrus.New())),
			),
		),
		grpc.StreamInterceptor(
			grpc_middleware.ChainStreamServer(
				otelgrpc.StreamServerInterceptor(),
				grpc_ctxtags.StreamServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)),
				grpc_logrus.StreamServerInterceptor(logrus.NewEntry(logrus.New())),
			),
		),
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
			PresenceClient: fleet.NewPresenceClient(conn),
			ClusterClient:  fleet.NewClusterClient(conn),
		}),
	)
	logrus.Debugf("auth server: listening on %s:%s", host, port)
	if err := srv.Serve(lis); err != nil {
		logrus.Fatal(err)
	}
}
