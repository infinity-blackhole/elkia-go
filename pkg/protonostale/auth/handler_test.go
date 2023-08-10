package auth

import (
	"bufio"
	"bytes"
	"context"
	"net"
	"testing"
	"testing/iotest"

	"go.shikanime.studio/elkia/internal/auth"
	"go.shikanime.studio/elkia/internal/auth/authtest"
	"go.shikanime.studio/elkia/internal/cluster"
	"go.shikanime.studio/elkia/internal/cluster/clustertest"
	"go.shikanime.studio/elkia/internal/presence"
	"go.shikanime.studio/elkia/internal/presence/presencetest"
	fleet "go.shikanime.studio/elkia/pkg/api/fleet/v1alpha1"
	"golang.org/x/sync/errgroup"
)

func TestHandlerServeNosTale(t *testing.T) {
	ctx := context.Background()
	wg := errgroup.Group{}
	fakePresence := presencetest.NewFakePresence(presence.MemoryPresenceServerConfig{
		Identities: map[uint32]*presence.Identity{
			1: {
				Username: "ricofo8350@otanhome.com",
				Password: "9hibwiwiG2e6Nr",
			},
		},
		Sessions: map[uint32]*fleet.Session{},
		Seed:     1,
	})
	wg.Go(fakePresence.Serve)
	fakePresenceClient, err := fakePresence.Dial(ctx)
	if err != nil {
		t.Fatal(err)
	}
	fakeCluster := clustertest.NewFakeCluster(cluster.MemoryClusterServerConfig{
		Members: []*fleet.Member{
			{
				Id:         "gateway-alpha",
				WorldId:    1,
				ChannelId:  1,
				Addresses:  []string{"127.0.0.1:4124"},
				Name:       "world-alpha",
				Population: 10,
				Capacity:   100,
			},
			{
				Id:         "gateway-beta",
				WorldId:    1,
				ChannelId:  2,
				Addresses:  []string{"127.0.0.1:4125"},
				Name:       "world-alpha",
				Population: 10,
				Capacity:   100,
			},
			{
				Id:         "gateway-gamma",
				WorldId:    1,
				ChannelId:  3,
				Addresses:  []string{"127.0.0.1:4126"},
				Name:       "world-alpha",
				Population: 10,
				Capacity:   100,
			},
			{
				Id:         "gateway-delta",
				WorldId:    2,
				ChannelId:  1,
				Addresses:  []string{"127.0.0.1:4127"},
				Name:       "world-beta",
				Population: 10,
				Capacity:   100,
			},
		},
	})
	wg.Go(fakeCluster.Serve)
	fakeClusterClient, err := fakeCluster.Dial(ctx)
	if err != nil {
		t.Fatal(err)
	}
	fakeAuth := authtest.NewFakeAuth(auth.ServerConfig{
		PresenceClient: fakePresenceClient,
		ClusterClient:  fakeClusterClient,
	})
	wg.Go(fakeAuth.Serve)
	fakeAuthClient, err := fakeAuth.Dial(ctx)
	if err != nil {
		t.Fatal(err)
	}
	handler := NewHandler(HandlerConfig{
		AuthClient: fakeAuthClient,
	})
	serverConn, clientConn := net.Pipe()
	clientWriter := bufio.NewWriter(
		iotest.NewWriteLogger(t.Name(), clientConn),
	)
	clientReader := bufio.NewReader(
		iotest.NewReadLogger(t.Name(), clientConn),
	)
	defer clientConn.Close()
	defer serverConn.Close()
	go handler.ServeNosTale(serverConn)
	input := []byte(
		"\x9c\xbb\x9f\x02\x05\x03\x05\xf2\xff\xff\x04\x05\xff\x06\x0a\xf2\xc0" +
			"\xb9\xaf\xbb\xb4\xbb\x0a\xff\x05\x02\x92\xbb\xc6\xb1\xbc\xba\xbb" +
			"\xbd\xb5\xfc\xaf\xbb\xbd\xf2\x01\x05\x01\xff\x01\x09\x05\x04\x90" +
			"\x0a\xff\x04\x03\x09\x05\x04\xff\x00\x06\x03\xff\x03\x01\x04\x05" +
			"\x09\xff\x03\x06\x03\x06\x04\x05\x09\x00\x06\x05\x03\x06\xff\x90" +
			"\x00\x01\x04\x96\x05\x00\xff\x94\x04\x05\x06\x0a\x95\x00\x03\x90" +
			"\x00\xf2\x02\x02\x04\x94\x03\x06\x06\x04\xd7\x02\xfc\x09\xfc\xff" +
			"\xfc\xff\x02\x0a\x04\xd8",
	)
	if _, err := clientWriter.Write(input); err != nil {
		t.Fatalf("Failed to write auth message: %v", err)
	}
	if err := clientWriter.Flush(); err != nil {
		t.Fatalf("Failed to flush message: %v", err)
	}
	result, err := clientReader.ReadBytes(0x19)
	if err != nil {
		t.Fatalf("Failed to read line bytes: %v", err)
	}
	expected := []byte(
		"\x5d\x82\x63\x74\x62\x63\x2f\x40\x3f\x48\x47\x40\x43\x41\x44\x40\x3f" +
			"\x2f\x40\x41\x46\x3d\x3f\x3d\x3f\x3d\x40\x49\x43\x40\x41\x43\x49" +
			"\x42\x49\x40\x3d\x40\x3d\x86\x7e\x81\x7b\x73\x3c\x70\x7b\x7f\x77" +
			"\x70\x2f\x40\x41\x46\x3d\x3f\x3d\x3f\x3d\x40\x49\x43\x40\x41\x44" +
			"\x49\x42\x49\x40\x3d\x41\x3d\x86\x7e\x81\x7b\x73\x3c\x70\x7b\x7f" +
			"\x77\x70\x2f\x40\x41\x46\x3d\x3f\x3d\x3f\x3d\x40\x49\x43\x40\x41" +
			"\x45\x49\x42\x49\x40\x3d\x42\x3d\x86\x7e\x81\x7b\x73\x3c\x70\x7b" +
			"\x7f\x77\x70\x2f\x40\x41\x46\x3d\x3f\x3d\x3f\x3d\x40\x49\x43\x40" +
			"\x41\x46\x49\x42\x49\x41\x3d\x40\x3d\x86\x7e\x81\x7b\x73\x3c\x71" +
			"\x74\x83\x70\x2f\x3c\x40\x49\x3c\x40\x49\x3c\x40\x49\x40\x3f\x3f" +
			"\x3f\x3f\x3d\x40\x3f\x3f\x3f\x3f\x3d\x40\x19",
	)
	if !bytes.Equal(expected, result) {
		t.Fatalf("Expected %v, got %v", expected, result)
	}
}
