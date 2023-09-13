package clustertest

import (
	"context"
	"net"

	"go.shikanime.studio/elkia/internal/cluster"
	fleetpb "go.shikanime.studio/elkia/pkg/api/fleet/v1alpha1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

func NewFakeCluster(c cluster.MemoryClusterServerConfig) *FakeCluster {
	return &FakeCluster{
		c:   c,
		lis: bufconn.Listen(1024 * 1024),
	}
}

type FakeCluster struct {
	c   cluster.MemoryClusterServerConfig
	lis *bufconn.Listener
}

func (f *FakeCluster) Serve() error {
	server := grpc.NewServer()
	fleetpb.RegisterClusterServer(
		server,
		cluster.NewMemoryClusterServer(f.c),
	)
	return server.Serve(f.lis)
}

func (f *FakeCluster) Dial(ctx context.Context) (fleetpb.ClusterClient, error) {
	conn, err := grpc.DialContext(
		ctx,
		"bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return f.lis.DialContext(ctx)
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}
	return fleetpb.NewClusterClient(conn), nil
}

func (f *FakeCluster) Close() error {
	return f.lis.Close()
}
