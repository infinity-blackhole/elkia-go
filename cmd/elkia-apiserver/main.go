package main

import (
	"net"

	"github.com/infinity-blackhole/elkia/internal/apiserver"
	fleetv1alha1pb "github.com/infinity-blackhole/elkia/pkg/api/fleet/v1alpha1"
	ory "github.com/ory/client-go"
	"github.com/spf13/pflag"
	etcd "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var (
	elkiaApiServerEndpoint string
	etcdEndpoints          []string
)

func init() {
	pflag.StringVar(
		&elkiaApiServerEndpoint,
		"elkia-api-server-endpoint",
		"/var/run/elkia.sock",
		"Elkia API Server endpoint",
	)
	pflag.StringSliceVar(
		&etcdEndpoints,
		"etcd-endpoints",
		[]string{"http://localhost:2379"},
		"etcd endpoints",
	)
}

func main() {
	clientset, err := NewKubernetesClientSet()
	if err != nil {
		panic(err)
	}
	lis, err := net.Listen("unix", elkiaApiServerEndpoint)
	if err != nil {
		panic(err)
	}
	sessionStore, err := NewSessionStoreClient()
	if err != nil {
		panic(err)
	}
	srv := grpc.NewServer()
	fleetv1alha1pb.RegisterFleetServiceServer(
		srv,
		apiserver.NewFleetService(apiserver.FleetServiceConfig{
			KubernetesClientSet: clientset,
			IdentityProvider:    NewIdentityProviderClient(),
			SessionStore:        sessionStore,
		}),
	)
	if err := srv.Serve(lis); err != nil {
		panic(err)
	}
}

func NewIdentityProviderClient() *apiserver.IdentityProvider {
	return apiserver.NewIdentityProviderService(&apiserver.IdentityProviderServiceConfig{
		OryClient: ory.NewAPIClient(ory.NewConfiguration()),
	})
}

func NewEtcdClient() (*etcd.Client, error) {
	return etcd.New(etcd.Config{
		Endpoints: etcdEndpoints,
	})
}

func NewSessionStoreClient() (*apiserver.SessionStore, error) {
	etcd, err := NewEtcdClient()
	if err != nil {
		return nil, err
	}
	return apiserver.NewSessionStoreClient(apiserver.SessionStoreConfig{
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
