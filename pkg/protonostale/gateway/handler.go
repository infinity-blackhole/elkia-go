package gateway

import (
	"context"
	"net"

	eventing "github.com/infinity-blackhole/elkia/pkg/api/eventing/v1alpha1"
	"github.com/infinity-blackhole/elkia/pkg/protonostale"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
)

var name = "github.com/infinity-blackhole/elkia/internal/gateway"

type HandlerConfig struct {
	GatewayClient eventing.GatewayClient
}

func NewHandler(cfg HandlerConfig) *Handler {
	return &Handler{
		gateway: cfg.GatewayClient,
	}
}

type Handler struct {
	gateway eventing.GatewayClient
}

func (h *Handler) ServeNosTale(c net.Conn) {
	ctx := context.Background()
	dec := NewSessionDecoder(c)
	var msg protonostale.SyncFrame
	if err := dec.Decode(&msg); err != nil {
		c.Close()
		return
	}
	logrus.Debugf("gateway: received sync frame: %v", msg.String())
	conn := h.newChannelConn(c, &msg)
	conn.serve(ctx)
}

func (h *Handler) newChannelConn(
	c net.Conn,
	msg *protonostale.SyncFrame,
) *conn {
	sequence := msg.GetSequence()
	code := msg.GetCode()
	return &conn{
		rwc:          c,
		gateway:      h.gateway,
		dec:          NewChannelDecoder(c, code),
		enc:          NewEncoder(c),
		Code:         code,
		lastSequence: sequence,
	}
}

type conn struct {
	rwc          net.Conn
	dec          *ChannelDecoder
	enc          *Encoder
	gateway      eventing.GatewayClient
	Code         uint32
	lastSequence uint32
}

func (c *conn) serve(ctx context.Context) {
	_, span := otel.Tracer(name).Start(ctx, "Handle Messages")
	defer span.End()
	logrus.Debugf("gateway: serving connection from %v", c.rwc.RemoteAddr())
	switch err := c.handleFrames(ctx).(type) {
	case *protonostale.Status:
		if err := c.enc.Encode(err); err != nil {
			logrus.Errorf("gateway: failed to send error: %v", err)
		}
	default:
		if err := c.enc.Encode(
			protonostale.NewStatus(eventing.Code_UNEXPECTED_ERROR),
		); err != nil {
			logrus.Errorf("gateway: failed to send error: %v", err)
		}
	}
}

func (c *conn) handleFrames(ctx context.Context) error {
	stream, err := c.gateway.ChannelInteract(ctx)
	if err != nil {
		return protonostale.NewStatus(eventing.Code_UNEXPECTED_ERROR)
	}
	logrus.Debugf("gateway: created auth handoff interact stream")
	if err := c.handleSync(stream); err != nil {
		return err
	}
	if err := c.handleIdentifier(stream); err != nil {
		return err
	}
	logrus.Debugf("gateway: sent sync frame")
	if err := c.handlePassword(stream); err != nil {
		return err
	}
	logrus.Debugf("gateway: sent login frame")
	for {
		var msg protonostale.ChannelInteractRequest
		if err := c.dec.Decode(&msg); err != nil {
			return protonostale.NewStatus(eventing.Code_UNEXPECTED_ERROR)
		}
		if err := stream.Send(&msg.ChannelInteractRequest); err != nil {
			return protonostale.NewStatus(eventing.Code_UNEXPECTED_ERROR)
		}
	}
}

func (c *conn) handleSync(stream eventing.Gateway_ChannelInteractClient) error {
	if err := stream.Send(&eventing.ChannelInteractRequest{
		Payload: &eventing.ChannelInteractRequest_SyncFrame{
			SyncFrame: &eventing.SyncFrame{
				Sequence: c.lastSequence,
				Code:     c.Code,
			},
		},
	}); err != nil {
		return protonostale.NewStatus(eventing.Code_UNEXPECTED_ERROR)
	}
	return nil
}

func (c *conn) handleIdentifier(stream eventing.Gateway_ChannelInteractClient) error {
	var msg protonostale.IdentifierFrame
	if err := c.dec.Decode(&msg); err != nil {
		return protonostale.NewStatus(eventing.Code_BAD_CASE)
	}
	logrus.Debugf("gateway: read identifier frame: %v", msg.String())
	if err := stream.Send(&eventing.ChannelInteractRequest{
		Payload: &eventing.ChannelInteractRequest_IdentifierFrame{
			IdentifierFrame: &msg.IdentifierFrame,
		},
	}); err != nil {
		return protonostale.NewStatus(eventing.Code_UNEXPECTED_ERROR)
	}
	return nil
}

func (c *conn) handlePassword(stream eventing.Gateway_ChannelInteractClient) error {
	var msg protonostale.PasswordFrame
	if err := c.dec.Decode(&msg); err != nil {
		return err
	}
	logrus.Debugf("gateway: read password frame: %v", msg.String())
	if err := stream.Send(&eventing.ChannelInteractRequest{
		Payload: &eventing.ChannelInteractRequest_PasswordFrame{
			PasswordFrame: &msg.PasswordFrame,
		},
	}); err != nil {
		return protonostale.NewStatus(eventing.Code_UNEXPECTED_ERROR)
	}
	return nil
}
