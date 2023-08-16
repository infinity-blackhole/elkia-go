package main

import (
	"fmt"
	"net"
	"os"

	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.shikanime.studio/elkia/internal/clients"
	"go.shikanime.studio/elkia/internal/cluster"
	"go.shikanime.studio/elkia/internal/presence"
	fleet "go.shikanime.studio/elkia/pkg/api/fleet/v1alpha1"
	"google.golang.org/grpc"

	_ "go.shikanime.studio/elkia/internal/monitoring"
)

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
			IdentityManager: presence.NewOryIdentityServer(presence.OryIdentityServerConfig{
				OryClient: oryClient,
			}),
			SessionManager: presence.NewRedisSessionServer(presence.RedisSessionServerConfig{
				RedisClient: redisClient,
			}),
		}),
	)
	myNs := os.Getenv("MY_NAMESPACE")
	if myNs == "" {
		myNs = "elkia"
	}
	fleet.RegisterClusterServer(
		srv,
		cluster.NewKubernetesClusterServer(cluster.KubernetesClusterServerConfig{
			Namespace:        myNs,
			KubernetesClient: kubeCs,
		}),
	)
	logrus.Debugf("fleetserver: serving grpc")
	if err := srv.Serve(lis); err != nil {
		logrus.Fatal(err)
	}
}
