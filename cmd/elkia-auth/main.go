package main

import (
	"fmt"
	"os"

	"github.com/infinity-blackhole/elkia/internal/auth"
	fleet "github.com/infinity-blackhole/elkia/pkg/api/fleet/v1alpha1"
	"github.com/infinity-blackhole/elkia/pkg/nostale"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

func init() {
	logrus.SetLevel(logrus.DebugLevel)
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
	)
	if err != nil {
		logrus.Fatal(err)
	}
	logrus.Debugf("auth: connected to fleet at %s", elkiaFleetEndpoint)
	host := os.Getenv("HOST")
	if host == "" {
		host = "localhost"
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "4123"
	}
	srv := nostale.NewServer(nostale.ServerConfig{
		Addr: fmt.Sprintf("%s:%s", host, port),
		Handler: auth.NewHandler(auth.HandlerConfig{
			FleetClient: fleet.NewFleetClient(conn),
		}),
	})
	if err != nil {
		logrus.Fatal(err)
	}
	logrus.Debugf("auth: listening on %s:%s", host, port)
	if err := srv.ListenAndServe(); err != nil {
		logrus.Fatal(err)
	}
}
