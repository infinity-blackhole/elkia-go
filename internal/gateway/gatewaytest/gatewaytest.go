package gatewaytest

import (
	"context"
	"net"

	"github.com/infinity-blackhole/elkia/internal/gateway"
	eventing "github.com/infinity-blackhole/elkia/pkg/api/eventing/v1alpha1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

func NewFakeGateway(c gateway.ServerConfig) *FakeGateway {
	return &FakeGateway{
		c:   c,
		lis: bufconn.Listen(1024 * 1024),
	}
}

type FakeGateway struct {
	c   gateway.ServerConfig
	lis *bufconn.Listener
}

func (f *FakeGateway) Serve() error {
	server := grpc.NewServer()
	eventing.RegisterGatewayServer(
		server,
		gateway.NewServer(f.c),
	)
	return server.Serve(f.lis)
}

func (f *FakeGateway) Dial(ctx context.Context) (eventing.GatewayClient, error) {
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
	return eventing.NewGatewayClient(conn), nil
}

func (f *FakeGateway) Close() error {
	return f.lis.Close()
}
