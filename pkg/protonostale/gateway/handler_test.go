package gateway

import (
	"bufio"
	"context"
	"net"
	"testing"
	"testing/iotest"

	"github.com/infinity-blackhole/elkia/internal/gateway"
	"github.com/infinity-blackhole/elkia/internal/gateway/gatewaytest"
	"github.com/infinity-blackhole/elkia/internal/presence"
	"github.com/infinity-blackhole/elkia/internal/presence/presencetest"
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
	wg.Go(func() error {
		handler.ServeNosTale(serverConn)
		return nil
	})
	if _, err := clientWriter.Write(
		[]byte("\x96\xa5\xaa\xe0\x4f\x0e"),
	); err != nil {
		t.Fatalf("Failed to write sync frame: %v", err)
	}
	if _, err := clientWriter.Write(
		[]byte(
			"\xc6\xe4\xcb\x91\x46\xcd\xd6\xdc\xd0\xd9\xd0\xc4\x07\xd4\x49\xff" +
				"\xd0\xcb\xde\xd1\xd7\xd0\xd2\xda\xc1\x70\x43\xdc\xd0\xd2\x3f",
		),
	); err != nil {
		t.Fatalf("Failed to write identifier frame: %v", err)
	}
	if _, err := clientWriter.Write(
		[]byte(
			"\xc7\xe4\xcb\xa1\x10\x48\xd7\xd6\xdd\xc8\xd6\xc8\xd6\xf8\xc1\xa0" +
				"\x41\xda\xc1\xe0\x42\xf1\xcd\x3f",
		),
	); err != nil {
		t.Fatalf("Failed to write identifier frame: %v", err)
	}
	if _, err := clientWriter.Write(
		[]byte("\xc7\xcd\xab\xf1\x80\x3f\x0a"),
	); err != nil {
		t.Fatalf("Failed to write heartbeat frame: %v", err)
	}
	if err := clientWriter.Flush(); err != nil {
		t.Fatalf("Failed to flush frame: %v", err)
	}
	wg.Wait()
}
