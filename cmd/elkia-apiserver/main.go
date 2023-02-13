package main

import (
	"net"

	"github.com/infinity-blackhole/elkia/internal/apiserver"
	"github.com/infinity-blackhole/elkia/internal/app"
	fleetv1alha1 "github.com/infinity-blackhole/elkia/pkg/api/fleet/v1alpha1"
	ory "github.com/ory/client-go"
	etcd "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func main() {
	c := new(app.Config)
	clientset, err := NewKubernetesClientSet(c)
	if err != nil {
		panic(err)
	}
	lis, err := net.Listen("unix", "/var/run/elkia.sock")
	if err != nil {
		panic(err)
	}
	sessionStore, err := NewSessionStoreClient(c)
	if err != nil {
		panic(err)
	}
	srv := grpc.NewServer()
	fleetv1alha1.RegisterFleetServiceServer(
		srv,
		apiserver.NewFleetService(apiserver.FleetServiceConfig{
			KubernetesClientSet: clientset,
			IdentityProvider:    NewIdentityProviderClient(c),
			SessionStore:        sessionStore,
		}),
	)
	if err := srv.Serve(lis); err != nil {
		panic(err)
	}
}

func NewIdentityProviderClient(c *app.Config) *apiserver.IdentityProvider {
	return apiserver.NewIdentityProviderService(&apiserver.IdentityProviderServiceConfig{
		OryClient: ory.NewAPIClient(ory.NewConfiguration()),
	})
}

func NewEtcdClient(c *app.Config) (*etcd.Client, error) {
	return etcd.New(etcd.Config{
		Endpoints: []string{"http://localhost:2379"},
	})
}

func NewSessionStoreClient(c *app.Config) (*apiserver.SessionStore, error) {
	etcd, err := NewEtcdClient(c)
	if err != nil {
		return nil, err
	}
	return apiserver.NewSessionStoreClient(apiserver.SessionStoreConfig{
		Etcd: etcd,
	}), nil
}

func NewKubernetesConfig(c *app.Config) (*rest.Config, error) {
	return rest.InClusterConfig()
}

func NewKubernetesClientSet(c *app.Config) (*kubernetes.Clientset, error) {
	config, err := NewKubernetesConfig(c)
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(config)
}
