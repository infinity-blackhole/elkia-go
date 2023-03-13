package main

import (
	"fmt"
	"os"

	eventing "github.com/infinity-blackhole/elkia/pkg/api/eventing/v1alpha1"
	"github.com/infinity-blackhole/elkia/pkg/nostale"
	"github.com/infinity-blackhole/elkia/pkg/protonostale/auth"
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
	endpoint := os.Getenv("ELKIA_AUTH_ENDPOINT")
	if endpoint == "" {
		endpoint = "localhost:8080"
	}
	logrus.Debugf("auth: connecting to auth at %s", endpoint)
	conn, err := grpc.Dial(
		endpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
		grpc.WithStreamInterceptor(otelgrpc.StreamClientInterceptor()),
	)
	if err != nil {
		logrus.Fatal(err)
	}
	logrus.Debugf("auth: connected to auth at %s", endpoint)
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
			AuthClient: eventing.NewAuthClient(conn),
		}),
	})
	logrus.Debugf("auth: listening on %s:%s", host, port)
	if err := srv.ListenAndServe(); err != nil {
		logrus.Fatal(err)
	}
}
