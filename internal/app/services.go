package app

import (
	core "github.com/infinity-blackhole/elkia/pkg/core"
	ory "github.com/ory/client-go"
	etcd "go.etcd.io/etcd/client/v3"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

func NewIdentityProviderClient(c *Config) *core.IdentityProvider {
	return core.NewIdentityProviderClient(&core.IdentityProviderClientConfig{
		OryClient: ory.NewAPIClient(ory.NewConfiguration()),
	})
}

func NewEtcdClient(c *Config) (*etcd.Client, error) {
	return etcd.New(etcd.Config{
		Endpoints: []string{"http://localhost:2379"},
	})
}

func NewSessionStoreClient(c *Config) (*core.SessionStore, error) {
	etcd, err := NewEtcdClient(c)
	if err != nil {
		return nil, err
	}
	return core.NewSessionStoreClient(core.SessionStoreConfig{
		Etcd: etcd,
	}), nil
}

func NewKubernetesConfig(c *Config) (*rest.Config, error) {
	return rest.InClusterConfig()
}

func NewKubernetesClientSet(c *Config) (*kubernetes.Clientset, error) {
	config, err := NewKubernetesConfig(c)
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(config)
}

func NewFleetClient(c *Config) (*core.FleetClient, error) {
	clientset, err := NewKubernetesClientSet(c)
	if err != nil {
		return nil, err
	}
	return core.NewFleetClient(core.FleetClientConfig{
		ClientSet: clientset,
	}), nil
}
