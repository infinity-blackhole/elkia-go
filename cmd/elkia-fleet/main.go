package main

import (
	"errors"
	"fmt"
	"net"
	"os"
	"strings"

	fleetserver "github.com/infinity-blackhole/elkia/internal/fleet"
	fleet "github.com/infinity-blackhole/elkia/pkg/api/fleet/v1alpha1"
	ory "github.com/ory/client-go"
	"github.com/sirupsen/logrus"
	etcd "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func init() {
	logrus.SetLevel(logrus.DebugLevel)
}

func main() {
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
	sessionStore, err := NewSessionStore()
	if err != nil {
		logrus.Fatal(err)
	}
	orchestrator, err := NewOrchestrator()
	if err != nil {
		logrus.Fatal(err)
	}
	srv := grpc.NewServer()
	fleet.RegisterFleetServer(
		srv,
		fleetserver.NewFleetServer(fleetserver.FleetServerConfig{
			Orchestrator:     orchestrator,
			IdentityProvider: NewIdentityProvider(),
			SessionStore:     sessionStore,
		}),
	)
	if err := srv.Serve(lis); err != nil {
		logrus.Fatal(err)
	}
}

func NewOrchestrator() (*fleetserver.Orchestrator, error) {
	clientset, err := NewKubernetesClientSet()
	if err != nil {
		return nil, err
	}
	return fleetserver.NewOrchestrator(fleetserver.OrchestratorConfig{
		KubernetesClientSet: clientset,
	}), nil
}

func NewIdentityProvider() *fleetserver.IdentityProvider {
	kratosEndpoint := os.Getenv("ELKIA_FLEET_KRATOS_ENDPOINT")
	if kratosEndpoint == "" {
		kratosEndpoint = "http://localhost:4433"
	}
	return fleetserver.NewIdentityProvider(&fleetserver.IdentityProviderServiceConfig{
		OryClient: ory.NewAPIClient(&ory.Configuration{
			DefaultHeader: make(map[string]string),
			UserAgent:     "OpenAPI-Generator/1.0.0/go",
			Debug:         false,
			Servers: ory.ServerConfigurations{
				{URL: kratosEndpoint},
			},
		}),
	})
}

func NewEtcd() (*etcd.Client, error) {
	etcdEndpointsStr := os.Getenv("ETCD_ENDPOINTS")
	var etcdEndpoints []string
	if etcdEndpointsStr == "" {
		etcdEndpoints = []string{"http://localhost:2379"}
	} else {
		etcdEndpoints = strings.Split(etcdEndpointsStr, ",")
	}
	etcdUsername := os.Getenv("ETCD_USERNAME")
	if etcdUsername == "" {
		etcdUsername = "root"
	}
	etcdPassword := os.Getenv("ETCD_PASSWORD")
	if etcdPassword == "" {
		return nil, errors.New("etcd password is required")
	}
	return etcd.New(etcd.Config{
		Endpoints: etcdEndpoints,
		Username:  etcdUsername,
		Password:  etcdPassword,
	})
}

func NewSessionStore() (*fleetserver.SessionStore, error) {
	etcd, err := NewEtcd()
	if err != nil {
		return nil, err
	}
	return fleetserver.NewSessionStoreClient(fleetserver.SessionStoreConfig{
		Etcd: etcd,
	}), nil
}

func NewKubernetesConfig() (*rest.Config, error) {
	return rest.InClusterConfig()
}

func NewKubernetesClientSet() (*kubernetes.Clientset, error) {
	config, err := NewKubernetesConfig()
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(config)
}
