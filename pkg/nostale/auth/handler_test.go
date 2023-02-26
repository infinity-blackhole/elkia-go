package auth

import (
	"bufio"
	"bytes"
	"context"
	"net"
	"testing"

	"github.com/infinity-blackhole/elkia/internal/cluster"
	"github.com/infinity-blackhole/elkia/internal/cluster/clustertest"
	"github.com/infinity-blackhole/elkia/internal/presence"
	"github.com/infinity-blackhole/elkia/internal/presence/presencetest"
	fleet "github.com/infinity-blackhole/elkia/pkg/api/fleet/v1alpha1"
	"golang.org/x/sync/errgroup"
)

func TestHandlerServeNosTale(t *testing.T) {
	ctx := context.Background()
	wg := errgroup.Group{}
	fakePresence := presencetest.NewFakePresence(presence.MemoryPresenceServerConfig{
		Identities: map[uint32]*presence.Identity{
			1: {
				Username: "admin",
				Password: "s3cr3t",
			},
		},
		Seed: 1,
	})
	wg.Go(fakePresence.Serve)
	_, err := fakePresence.Dial(ctx)
	if err != nil {
		t.Fatal(err)
	}
	fakeCluster := clustertest.NewFakeCluster(cluster.MemoryClusterServerConfig{
		Members: []*fleet.Member{
			{
				Id:         "gateway-alpha",
				WorldId:    1,
				ChannelId:  1,
				Address:    "127.0.0.1:4124",
				Name:       "world-alpha",
				Population: 10,
				Capacity:   100,
			},
			{
				Id:         "gateway-beta",
				WorldId:    1,
				ChannelId:  2,
				Address:    "127.0.0.1:4125",
				Name:       "world-alpha",
				Population: 10,
				Capacity:   100,
			},
			{
				Id:         "gateway-gamma",
				WorldId:    1,
				ChannelId:  3,
				Address:    "127.0.0.1:4126",
				Name:       "world-alpha",
				Population: 10,
				Capacity:   100,
			},
			{
				Id:         "gateway-delta",
				WorldId:    2,
				ChannelId:  1,
				Address:    "127.0.0.1:4127",
				Name:       "world-beta",
				Population: 10,
				Capacity:   100,
			},
		},
	})
	wg.Go(fakeCluster.Serve)
	_, err = fakeCluster.Dial(ctx)
	if err != nil {
		t.Fatal(err)
	}
	// TODO: Fix this test
	handler := NewHandler(HandlerConfig{})
	server, client := net.Pipe()
	defer client.Close()
	defer server.Close()
	handler.ServeNosTale(server)
	input := []byte{
		156, 187, 159, 2, 5, 3, 5, 242, 1, 2, 1, 5, 6, 4, 9, 9, 242, 177, 182,
		189, 185, 188, 242, 1, 1, 10, 6, 3, 255, 255, 1, 255, 5, 255, 255, 4,
		6, 255, 6, 3, 5, 0, 0, 255, 5, 255, 6, 3, 144, 6, 242, 2, 2, 255, 0,
		145, 2, 9, 2, 215, 2, 252, 9, 252, 255, 252, 255, 2, 10, 4, 216,
	}
	if _, err := client.Write(input); err != nil {
		t.Fatalf("Failed to write message: %v", err)
	}
	rc := bufio.NewReader(NewReader(bufio.NewReader(client)))
	result, err := rc.ReadBytes('\x0a')
	if err != nil {
		t.Fatalf("Failed to read line bytes: %v", err)
	}
	expected := []byte{
		141, 176, 151, 166, 144, 151, 227, 246, 240, 250, 251, 250, 245, 247,
		243, 242, 227, 242, 241, 244, 237, 243, 237, 243, 237, 242, 249, 247,
		242, 241, 247, 249, 240, 249, 242, 237, 242, 237, 180, 172, 177, 175,
		167, 238, 162, 175, 179, 171, 162, 227, 242, 241, 244, 237, 243, 237,
		243, 237, 242, 249, 247, 242, 241, 246, 249, 240, 249, 242, 237, 241,
		237, 180, 172, 177, 175, 167, 238, 162, 175, 179, 171, 162, 227, 242,
		241, 244, 237, 243, 237, 243, 237, 242, 249, 247, 242, 241, 245, 249,
		240, 249, 242, 237, 240, 237, 180, 172, 177, 175, 167, 238, 162, 175,
		179, 171, 162, 227, 242, 241, 244, 237, 243, 237, 243, 237, 242, 249,
		247, 242, 241, 244, 249, 240, 249, 241, 237, 242, 237, 180, 172, 177,
		175, 167, 238, 161, 166, 183, 162, 227, 238, 242, 249, 238, 242, 249,
		238, 242, 249, 242, 243, 243, 243, 243, 237, 242, 243, 243, 243, 243,
		237, 242, 201, 10,
	}
	if !bytes.Equal(result, expected) {
		t.Fatalf("Expected %v, got %v", expected, result)
	}
}
