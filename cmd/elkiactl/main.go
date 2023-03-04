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
	elkiaApiServerEndpoint := os.Getenv("ELKIA_APISERVER_ENDPOINT")
	if elkiaApiServerEndpoint == "" {
		elkiaApiServerEndpoint = "localhost:8080"
	}
	logrus.Debugf("auth: connecting to api server at %s", elkiaApiServerEndpoint)
	conn, err := grpc.Dial(
		elkiaApiServerEndpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		logrus.Fatal(err)
	}
	logrus.Debugf("auth: connected to api server at %s", elkiaApiServerEndpoint)
	fleet.NewPresenceClient(conn)
	fleet.NewClusterClient(conn)
}
