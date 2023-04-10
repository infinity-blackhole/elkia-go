package main

import (
	"os"

	fleet "github.com/infinity-blackhole/elkia/pkg/api/fleet/v1alpha1"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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
	)
	if err != nil {
		logrus.Fatal(err)
	}
	logrus.Debugf("auth: connected to fleet at %s", elkiaFleetEndpoint)
	fleet.NewPresenceClient(conn)
	fleet.NewClusterClient(conn)
}
