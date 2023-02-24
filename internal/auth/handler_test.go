package auth

import (
	"bufio"
	"context"
	"net"
	"testing"

	"github.com/infinity-blackhole/elkia/internal/cluster"
	"github.com/infinity-blackhole/elkia/internal/cluster/clustertest"
	"github.com/infinity-blackhole/elkia/internal/presence"
	"github.com/infinity-blackhole/elkia/internal/presence/presencetest"
	fleet "github.com/infinity-blackhole/elkia/pkg/api/fleet/v1alpha1"
	"github.com/infinity-blackhole/elkia/pkg/nostale/simplesubtitution"
	"golang.org/x/sync/errgroup"
)

func TestHandlerServeNosTale(t *testing.T) {
	ctx := context.Background()
	wg := errgroup.Group{}
	fakePresence := presencetest.NewFakePresence(presence.MemoryPresenceServerConfig{
		Identities: map[uint32]*presence.Identity{
			1: {
				Username: "admin",
				Password: "admin",
			},
			2: {
				Username: "user",
				Password: "user",
			},
		},
	})
	wg.Go(fakePresence.Serve)
	presenceClient, err := fakePresence.Dial(ctx)
	if err != nil {
		t.Fatal(err)
	}
	fakeCluster := clustertest.NewFakeCluster(cluster.MemoryClusterServerConfig{
		Members: []*fleet.Member{
			{
				Id:         "gateway-alpha",
				WorldId:    1,
				ChannelId:  1,
				Name:       "world-alpha",
				Population: 10,
				Capacity:   100,
			},
			{
				Id:         "gateway-beta",
				WorldId:    1,
				ChannelId:  2,
				Name:       "world-alpha",
				Population: 10,
				Capacity:   100,
			},
			{
				Id:         "gateway-gamma",
				WorldId:    1,
				ChannelId:  3,
				Name:       "world-alpha",
				Population: 10,
				Capacity:   100,
			},
			{
				Id:         "gateway-delta",
				WorldId:    2,
				ChannelId:  1,
				Name:       "world-beta",
				Population: 10,
				Capacity:   100,
			},
		},
	})
	wg.Go(fakeCluster.Serve)
	clusterClient, err := fakeCluster.Dial(ctx)
	if err != nil {
		t.Fatal(err)
	}
	handler := NewHandler(HandlerConfig{
		PresenceClient: presenceClient,
		ClusterClient:  clusterClient,
	})
	server, client := net.Pipe()
	defer client.Close()
	defer server.Close()
	handler.ServeNosTale(server)
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
