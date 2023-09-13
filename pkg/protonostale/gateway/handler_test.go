package gateway

import (
	"bufio"
	"context"
	"net"
	"testing"
	"testing/iotest"

	"go.shikanime.studio/elkia/internal/gateway"
	"go.shikanime.studio/elkia/internal/gateway/gatewaytest"
	"go.shikanime.studio/elkia/internal/presence"
	"go.shikanime.studio/elkia/internal/presence/presencetest"
	fleetpb "go.shikanime.studio/elkia/pkg/api/fleet/v1alpha1"
	"golang.org/x/sync/errgroup"
)

// TODO: For the moment we have no idea what is the logout frame
func TestHandlerServeNosTale(t *testing.T) {
	ctx := context.Background()
	var wg errgroup.Group
	fakePresence := presencetest.NewFakePresence(presence.MemoryPresenceServerConfig{
		Identities: map[uint32]*presence.Identity{
			1: {
				Username: "ricofo8350@otanhome.com",
				Password: "9hibwiwiG2e6Nr",
			},
		},
		Sessions: map[uint32]*fleetpb.Session{
			0: {
				Token: "zIRkVceCdjQXzB1feO9Sukm2N8dPzS3TQ3mI9GyZ0Z1EVpmGFI3HKat114vreNOh",
			},
		},
		Seed: 1,
	})
	wg.Go(fakePresence.Serve)
	fakePresenceClient, err := fakePresence.Dial(ctx)
	if err != nil {
		t.Fatal(err)
	}
	fakeGateway := gatewaytest.NewFakeGateway(gateway.ServerConfig{
		PresenceClient: fakePresenceClient,
	})
	wg.Go(fakeGateway.Serve)
	fakeGatewayClient, err := fakeGateway.Dial(ctx)
	if err != nil {
		t.Fatal(err)
	}
	handler := NewHandler(HandlerConfig{
		GatewayClient: fakeGatewayClient,
	})
	serverConn, clientConn := net.Pipe()
	clientWriter := bufio.NewWriter(
		iotest.NewWriteLogger(t.Name(), clientConn),
	)
	defer clientConn.Close()
	defer serverConn.Close()
	go handler.ServeNosTale(serverConn)
	if _, err := clientWriter.Write(
		[]byte("\x96\x94\xa9\xe0\x4f\x0e"),
	); err != nil {
		t.Fatalf("Failed to write sync frame: %v", err)
	}
	if _, err := clientWriter.Write(
		[]byte(
			"\xc6\xc5\xdb\x81\x46\xcd\xd6\xdc\xd0\xd9\xd0\xc4\x07\xd4\x49\xff" +
				"\xd0\xcb\xde\xd1\xd7\xd0\xd2\xda\xc1\x70\x43\xdc\xd0\xd2\x3f",
		),
	); err != nil {
		t.Fatalf("Failed to write identifier frame: %v", err)
	}
	if _, err := clientWriter.Write(
		[]byte(
			"\xc7\xc5\xdb\x91\x10\x48\xd7\xd6\xdd\xc8\xd6\xc8\xd6\xf8\xc1\xa0" +
				"\x41\xda\xc1\xe0\x42\xf1\xcd\x3f",
		),
	); err != nil {
		t.Fatalf("Failed to write identifier frame: %v", err)
	}
	if _, err := clientWriter.Write(
		[]byte("\xc7\xc5\xdb\xa1\x80\x3f"),
	); err != nil {
		t.Fatalf("Failed to write heartbeat frame: %v", err)
	}
	if _, err := clientWriter.Write(
		[]byte("\xc7\xc5\xdb\xb1\x80\x3f"),
	); err != nil {
		t.Fatalf("Failed to write heartbeat frame: %v", err)
	}
	if err := clientWriter.Flush(); err != nil {
		t.Fatalf("Failed to flush frame: %v", err)
	}
	if err := clientConn.Close(); err != nil {
		t.Fatalf("Failed to close client connection: %v", err)
	}
}
