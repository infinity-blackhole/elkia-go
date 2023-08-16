package presencetest

import (
	"context"
	"net"

	"go.shikanime.studio/elkia/internal/presence"
	fleet "go.shikanime.studio/elkia/pkg/api/fleet/v1alpha1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

func NewFakePresence(c presence.PresenceServerConfig) *FakePresence {
	return &FakePresence{
		c:   c,
		lis: bufconn.Listen(1024 * 1024),
	}
}

type FakePresence struct {
	c   presence.PresenceServerConfig
	lis *bufconn.Listener
}

func (f *FakePresence) Serve() error {
	server := grpc.NewServer()
	fleet.RegisterPresenceServer(
		server,
		presence.NewPresenceServer(f.c),
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

func (f *FakePresence) Close() error {
	return f.lis.Close()
}
