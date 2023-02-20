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
	sessionStore, err := NewSessionStore()
	if err != nil {
		logrus.Fatal(err)
	}
	logrus.Debugf("fleetserver: connected to etcd")
	orchestrator, err := NewOrchestrator()
	if err != nil {
		logrus.Fatal(err)
	}
	logrus.Debugf("fleetserver: connected to kubernetes")
	srv := grpc.NewServer()
	fleet.RegisterFleetServer(
		srv,
		fleetserver.NewFleetServer(fleetserver.FleetServerConfig{
			Orchestrator:     orchestrator,
			IdentityProvider: NewIdentityProvider(),
			SessionStore:     sessionStore,
		}),
	)
	logrus.Debugf("fleetserver: serving grpc")
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
	kratosUrlStr := os.Getenv("KRATOS_URIS")
	var kratosUrls []string
	if kratosUrlStr != "" {
		kratosUrls = strings.Split(kratosUrlStr, ",")
	} else {
		kratosUrls = []string{"http://localhost:4433"}
	}
	var oryServerConfigs []ory.ServerConfiguration
	for _, url := range kratosUrls {
		oryServerConfigs = append(oryServerConfigs, ory.ServerConfiguration{URL: url})
	}
	return fleetserver.NewIdentityProvider(&fleetserver.IdentityProviderServiceConfig{
		OryClient: ory.NewAPIClient(&ory.Configuration{
			DefaultHeader: make(map[string]string),
			UserAgent:     "OpenAPI-Generator/1.0.0/go",
			Debug:         false,
			Servers:       oryServerConfigs,
		}),
	})
}

func NewEtcd() (*etcd.Client, error) {
	etcdUrisStr := os.Getenv("ETCD_URIS")
	var etcdUris []string
	if etcdUrisStr == "" {
		etcdUris = []string{"http://localhost:2379"}
	} else {
		etcdUris = strings.Split(etcdUrisStr, ",")
	}
	etcdUsername := os.Getenv("ETCD_USERNAME")
	if etcdUsername == "" {
		etcdUsername = "root"
	}
	etcdPassword := os.Getenv("ETCD_PASSWORD")
	if etcdPassword == "" {
		return nil, errors.New("etcd password is required")
	}
	logrus.Debugf("fleet server connecting to etcd: %s", etcdUris)
	return etcd.New(etcd.Config{
		Endpoints: etcdUris,
		Username:  etcdUsername,
		Password:  etcdPassword,
	})
}

func NewSessionStore() (*fleetserver.SessionStore, error) {
	cli, err := NewEtcd()
	if err != nil {
		return nil, err
	}
	return fleetserver.NewSessionStoreClient(fleetserver.SessionStoreConfig{
		Etcd: etcd.NewKV(cli),
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
