package main

import (
	"net"

	fleetserver "github.com/infinity-blackhole/elkia/internal/fleet"
	fleet "github.com/infinity-blackhole/elkia/pkg/api/fleet/v1alpha1"
	ory "github.com/ory/client-go"
	"github.com/spf13/pflag"
	etcd "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var (
	address        string
	etcdEndpoints  []string
	etchUsername   string
	etchPassword   string
	kratosEndpoint string
)

func init() {
	pflag.StringVar(
		&address,
		"address",
		":8080",
		"Address",
	)
	pflag.StringSliceVar(
		&etcdEndpoints,
		"etcd-endpoints",
		[]string{"http://localhost:2379"},
		"Etcd endpoints",
	)
	pflag.StringVar(
		&etchUsername,
		"etcd-username",
		"",
		"Etcd username",
	)
	pflag.StringVar(
		&etchPassword,
		"etcd-password",
		"",
		"Etcd password",
	)
	pflag.StringVar(
		&kratosEndpoint,
		"kratos-endpoint",
		"http://localhost:4433",
		"Kratos endpoint",
	)
}

func main() {
	pflag.Parse()
	lis, err := net.Listen("tcp", address)
	if err != nil {
		panic(err)
	}
	sessionStore, err := NewSessionStore()
	if err != nil {
		panic(err)
	}
	orchestrator, err := NewOrchestrator()
	if err != nil {
		panic(err)
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
		panic(err)
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
	return etcd.New(etcd.Config{
		Endpoints: etcdEndpoints,
		Username:  etchUsername,
		Password:  etchPassword,
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
