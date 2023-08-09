package main

import (
	"fmt"
	"os"

	"github.com/infinity-blackhole/elkia/internal/clients"
	"github.com/infinity-blackhole/elkia/pkg/nostale"
	"github.com/infinity-blackhole/elkia/pkg/protonostale/auth"
	"github.com/sirupsen/logrus"

	_ "github.com/infinity-blackhole/elkia/internal/monitoring"
)

func main() {
	authCs, err := clients.NewAuthClientSet()
	if err != nil {
		logrus.Fatal(err)
	}
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
			AuthClient: authCs.AuthClient,
		}),
	})
	logrus.Debugf("auth: listening on %s:%s", host, port)
	if err := srv.ListenAndServe(); err != nil {
		logrus.Fatal(err)
	}
}
