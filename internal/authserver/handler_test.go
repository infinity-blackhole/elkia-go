package authserver

import (
	"bufio"
	"context"
	"crypto/sha1"
	"encoding/base64"
	"hash/fnv"
	"net"
	"testing"

	fleetv1alpha1pb "github.com/infinity-blackhole/elkia/pkg/api/fleet/v1alpha1"
	"github.com/infinity-blackhole/elkia/pkg/nostale/simplesubtitution"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

type fleetServiceServerMock struct {
	fleetv1alpha1pb.UnimplementedFleetServiceServer
}

func (n *fleetServiceServerMock) CreateHandoff(
	ctx context.Context,
	in *fleetv1alpha1pb.CreateHandoffRequest,
) (*fleetv1alpha1pb.CreateHandoffResponse, error) {
	h := fnv.New32a()
	h.Write([]byte(in.Identifier))
	h.Write([]byte(in.Token))
	key := h.Sum32()
	return &fleetv1alpha1pb.CreateHandoffResponse{
		Key: key,
	}, nil
}

func (s *fleetServiceServerMock) ListClusters(
	ctx context.Context,
	in *fleetv1alpha1pb.ListClusterRequest,
) (*fleetv1alpha1pb.ListClusterResponse, error) {
	return &fleetv1alpha1pb.ListClusterResponse{
		Clusters: []*fleetv1alpha1pb.Cluster{
			{
				Id:      "foo",
				WorldId: 1,
				Name:    "test-1",
			},
			{
				Id:      "bar",
				WorldId: 2,
				Name:    "test-2",
			},
		},
	}, nil
}

func (s *fleetServiceServerMock) ListGateways(
	ctx context.Context,
	in *fleetv1alpha1pb.ListGatewayRequest,
) (*fleetv1alpha1pb.ListGatewayResponse, error) {
	sh := sha1.New()
	sh.Write([]byte(in.Id))
	id := base64.URLEncoding.EncodeToString(sh.Sum(nil))
	nh := fnv.New32a()
	nh.Write([]byte(in.Id))
	channelId := nh.Sum32()
	return &fleetv1alpha1pb.ListGatewayResponse{
		Gateways: []*fleetv1alpha1pb.Gateway{
			{
				Id:         id,
				ChannelId:  channelId,
				Address:    "127.0.0.1:4321",
				Population: 0,
				Capacity:   1000,
			},
		},
	}, nil
}

func serveFleetServiceServerMock(lis net.Listener) error {
	server := grpc.NewServer()
	fleetv1alpha1pb.RegisterFleetServiceServer(
		server, &fleetServiceServerMock{},
	)
	return server.Serve(lis)
}

func dialFleetServiceServerMock(
	ctx context.Context, lis *bufconn.Listener,
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

func TestHandlerServeNosTale(t *testing.T) {
	ctx := context.Background()
	wg := errgroup.Group{}
	lis := bufconn.Listen(1024 * 1024)
	wg.Go(func() error {
		return serveFleetServiceServerMock(lis)
	})
	server, client := net.Pipe()
	defer client.Close()
	wg.Go(func() error {
		fleetClient, err := dialFleetServiceServerMock(ctx, lis)
		if err != nil {
			return err
		}
		handler := NewHandler(HandlerConfig{
			FleetClient: fleetClient,
		})
		handler.ServeNosTale(server)
		return server.Close()
	})
	input := []byte{
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
	}
	if _, err := client.Write(input); err != nil {
		t.Fatalf("Failed to write message: %v", err)
	}
	rc := simplesubtitution.NewReader(bufio.NewReader(client))
	result, err := rc.ReadMessageBytes()
	if err != nil {
		t.Fatalf("Failed to read line bytes: %v", err)
	}
	if len(result) == 0 {
		t.Fatalf("Empty message")
	}
}