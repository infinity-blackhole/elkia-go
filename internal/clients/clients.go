package clients

import (
	"os"
	"strings"

	eventing "github.com/infinity-blackhole/elkia/pkg/api/eventing/v1alpha1"
	fleet "github.com/infinity-blackhole/elkia/pkg/api/fleet/v1alpha1"
	ory "github.com/ory/client-go"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

func NewOryClient() *ory.APIClient {
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
	return ory.NewAPIClient(&ory.Configuration{
		DefaultHeader: make(map[string]string),
		UserAgent:     "OpenAPI-Generator/1.0.0/go",
		Debug:         false,
		Servers:       oryServerConfigs,
	})
}

func NewRedisClient() redis.UniversalClient {
	redisAddrsStr := os.Getenv("REDIS_ADDRESSES")
	var redisUrls []string
	if redisAddrsStr != "" {
		redisUrls = strings.Split(redisAddrsStr, ",")
	} else {
		redisUrls = []string{"localhost:6379"}
	}
	redisPwd := os.Getenv("REDIS_PASSWORD")
	redisClient := redis.NewUniversalClient(&redis.UniversalOptions{
		Addrs:    redisUrls,
		Password: redisPwd,
	})
	return redisClient
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

type FleetClientSet struct {
	PresenceClient fleet.PresenceClient
	ClusterClient  fleet.ClusterClient
}

func NewFleetClientSet() (*FleetClientSet, error) {
	fleetEndpoint := os.Getenv("ELKIA_FLEET_ENDPOINT")
	if fleetEndpoint == "" {
		fleetEndpoint = "localhost:8080"
	}
	logrus.Debugf("gateway: connecting to fleet at %s", fleetEndpoint)
	fleetConn, err := grpc.Dial(
		fleetEndpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
		grpc.WithStreamInterceptor(otelgrpc.StreamClientInterceptor()),
	)
	logrus.Debugf("gateway: connected to fleet at %s", fleetEndpoint)
	return &FleetClientSet{
		PresenceClient: fleet.NewPresenceClient(fleetConn),
		ClusterClient:  fleet.NewClusterClient(fleetConn),
	}, err
}

type GatewayClientSet struct {
	GatewayClient eventing.GatewayClient
}

func NewGatewayClientSet() (*GatewayClientSet, error) {
	gatewayEndpoint := os.Getenv("ELKIA_GATEWAY_ENDPOINT")
	if gatewayEndpoint == "" {
		gatewayEndpoint = "localhost:8081"
	}
	logrus.Debugf("gateway: connecting to gateway at %s", gatewayEndpoint)
	gatewayConn, err := grpc.Dial(
		gatewayEndpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
		grpc.WithStreamInterceptor(otelgrpc.StreamClientInterceptor()),
	)
	logrus.Debugf("gateway: connected to gateway at %s", gatewayEndpoint)
	return &GatewayClientSet{
		GatewayClient: eventing.NewGatewayClient(gatewayConn),
	}, err
}

type AuthClientSet struct {
	AuthClient eventing.AuthClient
}

func NewAuthClientSet() (*AuthClientSet, error) {
	authEndpoint := os.Getenv("ELKIA_AUTH_ENDPOINT")
	if authEndpoint == "" {
		authEndpoint = "localhost:8082"
	}
	logrus.Debugf("gateway: connecting to auth at %s", authEndpoint)
	authConn, err := grpc.Dial(
		authEndpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
		grpc.WithStreamInterceptor(otelgrpc.StreamClientInterceptor()),
	)
	logrus.Debugf("gateway: connected to auth at %s", authEndpoint)
	return &AuthClientSet{
		AuthClient: eventing.NewAuthClient(authConn),
	}, err
}
