package main

import (
	"fmt"
	"net"
	"os"

	"github.com/infinity-blackhole/elkia/internal/clients"
	"github.com/infinity-blackhole/elkia/internal/cluster"
	"github.com/infinity-blackhole/elkia/internal/presence"
	fleet "github.com/infinity-blackhole/elkia/pkg/api/fleet/v1alpha1"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
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
	logrus.Debugf("Starting fleet server")
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
	logrus.Debugf("fleetserver: listening on %s:%s", host, port)
	oryClient := clients.NewOryClient()
	logrus.Debugf("fleetserver: connected to ory")
	redisClient := clients.NewRedisClient()
	logrus.Debugf("fleetserver: connected to redis")
	kubeCs, err := clients.NewKubernetesClientSet()
	if err != nil {
		logrus.Fatal(err)
	}
	logrus.Debugf("fleetserver: connected to kubernetes")
	srv := grpc.NewServer(
		grpc.UnaryInterceptor(otelgrpc.UnaryServerInterceptor()),
		grpc.StreamInterceptor(otelgrpc.StreamServerInterceptor()),
	)
	fleet.RegisterPresenceServer(
		srv,
		presence.NewPresenceServer(presence.PresenceServerConfig{
			OryClient:   oryClient,
			RedisClient: redisClient,
		}),
	)
	fleet.RegisterClusterServer(
		srv,
		cluster.NewKubernetesClusterServer(cluster.KubernetesClusterServerConfig{
			Namespace:        "elkia",
			KubernetesClient: kubeCs,
		}),
	)
	logrus.Debugf("fleetserver: serving grpc")
	if err := srv.Serve(lis); err != nil {
		logrus.Fatal(err)
	}
}
