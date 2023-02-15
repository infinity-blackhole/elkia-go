package authserver

import (
	"bufio"
	"context"
	"net"
	"testing"

	fleetv1alpha1pb "github.com/infinity-blackhole/elkia/pkg/api/fleet/v1alpha1"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

type mockFleetServer struct {
	fleetv1alpha1pb.UnimplementedFleetServiceServer
}

func newFleetServer(s fleetv1alpha1pb.FleetServiceServer) *grpc.Server {
	server := grpc.NewServer()
	fleetv1alpha1pb.RegisterFleetServiceServer(server, s)
	return server
}

func serveFleetServer(
	lis *bufconn.Listener,
	s fleetv1alpha1pb.FleetServiceServer,
) error {
	return newFleetServer(s).Serve(lis)
}

func dialFleetService(
	ctx context.Context,
	lis *bufconn.Listener,
) (fleetv1alpha1pb.FleetServiceClient, error) {
	conn, err := grpc.DialContext(
		ctx,
		"bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return lis.DialContext(ctx)
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}
	return fleetv1alpha1pb.NewFleetServiceClient(conn), nil
}

func TestServeNosTaleCredentialMessage(t *testing.T) {
	ctx := context.Background()
	wg := errgroup.Group{}
	lis := bufconn.Listen(1024 * 1024)
	wg.Go(func() error {
		return serveFleetServer(lis, &mockFleetServer{})
	})
	fleetClient, err := dialFleetService(ctx, lis)
	if err != nil {
		t.Fatalf("Failed to dial fleet service: %v", err)
	}
	handler := NewHandler(HandlerConfig{
		FleetClient: fleetClient,
	})
	server, client := net.Pipe()
	defer client.Close()
	wg.Go(func() error {
		handler.ServeNosTale(server)
		return server.Close()
	})
	connWrite := bufio.NewWriter(client)
	if _, err := connWrite.Write([]byte{
		156, 187, 159, 2, 5, 3, 5, 242, 255, 4, 1, 6, 2, 255, 10, 242, 177,
		242, 5, 145, 149, 4, 0, 5, 4, 4, 5, 148, 255, 149, 2, 144, 150, 2, 145,
		2, 4, 5, 149, 150, 2, 3, 145, 6, 1, 9, 10, 9, 149, 6, 2, 0, 5, 144, 3,
		9, 150, 1, 255, 9, 255, 2, 145, 0, 145, 10, 143, 5, 3, 150, 4, 144, 6,
		255, 0, 5, 0, 0, 4, 3, 2, 3, 150, 9, 5, 4, 145, 2, 10, 0, 150, 1, 149,
		9, 1, 144, 6, 150, 9, 4, 145, 3, 9, 255, 5, 4, 0, 150, 148, 9, 10,
		148, 150, 2, 255, 143, 9, 150, 143, 148, 3, 6, 255, 143, 9, 143, 3,
		144, 6, 149, 255, 2, 5, 5, 150, 6, 148, 9, 148, 2, 9, 144, 145, 2, 1,
		5, 242, 2, 2, 255, 9, 149, 255, 150, 143, 215, 2, 252, 9, 252, 255,
		252, 255, 2, 3, 1, 242, 2, 242, 143, 3, 150, 0, 5, 2, 255, 144, 150,
		0, 5, 3, 148, 5, 144, 145, 149, 2, 10, 3, 2, 148, 6, 2, 143, 0, 150,
		145, 255, 4, 4, 4, 216,
	}); err != nil {
		t.Fatalf("Failed to write message: %v", err)
	}
	if err := connWrite.Flush(); err != nil {
		t.Fatalf("Failed to flush message: %v", err)
	}
	if err := wg.Wait(); err != nil {
		t.Fatalf("Failed to wait group: %v", err)
	}
}
