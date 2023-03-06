package gateway

import (
	"bufio"
	"context"
	"net"
	"testing"

	"github.com/infinity-blackhole/elkia/internal/gateway"
	"github.com/infinity-blackhole/elkia/internal/gateway/gatewaytest"
	"github.com/infinity-blackhole/elkia/internal/presence"
	"github.com/infinity-blackhole/elkia/internal/presence/presencetest"
	"github.com/infinity-blackhole/elkia/pkg/nostale/utils"
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
		utils.NewWriteLogger(t.Name(), clientConn),
	)
	defer clientConn.Close()
	defer serverConn.Close()
	wg.Go(func() error {
		handler.ServeNosTale(serverConn)
		return nil
	})
	syncInput := []byte{
		158, 166, 180, 192, 197, 132, 199, 230, 143, 14,
	}
	if _, err := clientWriter.Write(syncInput); err != nil {
		t.Fatalf("Failed to write message: %v", err)
	}
	handoffInput := []byte{
		198, 228, 203, 145, 70, 205, 214, 220, 208, 217, 208, 196, 7, 212, 73,
		255, 208, 203, 222, 209, 215, 208, 210, 218, 193, 112, 67, 220, 208,
		210, 63, 199, 228, 203, 161, 16, 72, 215, 214, 221, 200, 214, 200, 214,
		248, 193, 160, 65, 218, 193, 224, 66, 241, 205, 63, 10,
	}
	if _, err := clientWriter.Write(handoffInput); err != nil {
		t.Fatalf("Failed to write message: %v", err)
	}
	heartbeatInput := []byte{
		199, 205, 171, 241, 128, 63, 10,
	}
	if _, err := clientWriter.Write(heartbeatInput); err != nil {
		t.Fatalf("Failed to write message: %v", err)
	}
	if err := clientWriter.Flush(); err != nil {
		t.Fatalf("Failed to flush message: %v", err)
	}
	wg.Wait()
}
