package client

import (
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	eventing "go.shikanime.studio/elkia/pkg/api/eventing/v1alpha1"
	fleet "go.shikanime.studio/elkia/pkg/api/fleet/v1alpha1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

type ClientSet struct {
	PresenceClient fleet.PresenceClient
	ClusterClient  fleet.ClusterClient
	GatewayClient  eventing.GatewayClient
	AuthClient     eventing.AuthClient
}

func New(target string, cfg *Config) (*ClientSet, error) {
	conn, err := grpc.NewClient(
		target,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	)
	return &ClientSet{
		PresenceClient: fleet.NewPresenceClient(conn),
		ClusterClient:  fleet.NewClusterClient(conn),
		AuthClient:     eventing.NewAuthClient(conn),
		GatewayClient:  eventing.NewGatewayClient(conn),
	}, err
}
