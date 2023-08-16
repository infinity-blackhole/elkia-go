package main

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"go.shikanime.studio/elkia/internal/clients"
	"go.shikanime.studio/elkia/pkg/nostale"
	"go.shikanime.studio/elkia/pkg/protonostale/auth"

	_ "go.shikanime.studio/elkia/internal/monitoring"
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
		port = "4000"
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
