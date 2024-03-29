package main

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"go.shikanime.studio/elkia/internal/clients"
	"go.shikanime.studio/elkia/pkg/nostale"
	"go.shikanime.studio/elkia/pkg/protonostale/gateway"

	_ "go.shikanime.studio/elkia/internal/monitoring"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

func main() {
	gatewayClientSet, err := clients.NewGatewayClientSet()
	if err != nil {
		logrus.Fatal(err)
	}
	host := os.Getenv("HOST")
	if host == "" {
		host = "localhost"
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "4124"
	}
	srv := nostale.NewServer(nostale.ServerConfig{
		Addr: fmt.Sprintf("%s:%s", host, port),
		Handler: gateway.NewHandler(gateway.HandlerConfig{
			GatewayClient: gatewayClientSet.GatewayClient,
		}),
	})
	logrus.Debugf("auth: listening on %s:%s", host, port)
	if err := srv.ListenAndServe(); err != nil {
		logrus.Fatal(err)
	}
}
