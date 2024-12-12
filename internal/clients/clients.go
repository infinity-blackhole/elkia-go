package clients

import (
	"errors"
	"os"
	"strings"

	ory "github.com/ory/client-go"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	eventing "go.shikanime.studio/elkia/pkg/api/eventing/v1alpha1"
	fleet "go.shikanime.studio/elkia/pkg/api/fleet/v1alpha1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

func NewGormDialector() (gorm.Dialector, error) {
	dsn := os.Getenv("DSN")
	parts := strings.SplitN(dsn, "://", 2)
	switch parts[0] {
	case "mysql://":
		return mysql.Open(parts[1]), nil
	default:
		return nil, errors.New("unsupported database dialect")
	}
}

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
	redisEndpointsStr := os.Getenv("REDIS_ENDPOINTS")
	var redisEndpoints []string
	if redisEndpointsStr != "" {
		redisEndpoints = strings.Split(redisEndpointsStr, ",")
	} else {
		redisEndpoints = []string{"localhost:6379"}
	}
	redisUser := os.Getenv("REDIS_USERNAME")
	redisPwd := os.Getenv("REDIS_PASSWORD")
	redisClient := redis.NewUniversalClient(&redis.UniversalOptions{
		Addrs:    redisEndpoints,
		Username: redisUser,
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

type ClientSet struct {
	PresenceClient fleet.PresenceClient
	ClusterClient  fleet.ClusterClient
}

func NewClientSet() (*ClientSet, error) {
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
	return &ClientSet{
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
