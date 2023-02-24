package presencetest

import (
	"context"
	"net"

	"github.com/infinity-blackhole/elkia/internal/presence"
	fleet "github.com/infinity-blackhole/elkia/pkg/api/fleet/v1alpha1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

func NewFakePresence(c presence.MemoryPresenceServerConfig) *FakePresence {
	return &FakePresence{
		c:   c,
		lis: bufconn.Listen(1024 * 1024),
	}
}

type FakePresence struct {
	c   presence.MemoryPresenceServerConfig
	lis *bufconn.Listener
}

func (f *FakePresence) Serve() error {
	server := grpc.NewServer()
	fleet.RegisterPresenceServer(
		server,
		presence.NewMemoryPresenceServer(f.c),
	)
	return server.Serve(f.lis)
}

func (f *FakePresence) Dial(ctx context.Context) (fleet.PresenceClient, error) {
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
	return fleet.NewPresenceClient(conn), nil
}
