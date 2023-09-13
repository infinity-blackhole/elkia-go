package authtest

import (
	"context"
	"net"

	"go.shikanime.studio/elkia/internal/auth"
	eventingpb "go.shikanime.studio/elkia/pkg/api/eventing/v1alpha1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

func NewFakeAuth(c auth.ServerConfig) *FakeAuth {
	return &FakeAuth{
		c:   c,
		lis: bufconn.Listen(1024 * 1024),
	}
}

type FakeAuth struct {
	c   auth.ServerConfig
	lis *bufconn.Listener
}

func (f *FakeAuth) Serve() error {
	server := grpc.NewServer()
	eventingpb.RegisterAuthServer(
		server,
		auth.NewServer(f.c),
	)
	return server.Serve(f.lis)
}

func (f *FakeAuth) Dial(ctx context.Context) (eventingpb.AuthClient, error) {
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
	return eventingpb.NewAuthClient(conn), nil
}
