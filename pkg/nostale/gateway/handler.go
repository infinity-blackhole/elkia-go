package gateway

import (
	"context"
	"net"

	eventing "github.com/infinity-blackhole/elkia/pkg/api/eventing/v1alpha1"
	"github.com/infinity-blackhole/elkia/pkg/nostale/encoding"
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
	dec := encoding.NewDecoder(encoding.NewSessionReader(c))
	var sync protonostale.SyncFrame
	if err := dec.Decode(&sync); err != nil {
		c.Close()
		return
	}
	logrus.Debugf("gateway: received sync frame: %v", sync.String())
	conn := h.newChannelConn(c, &sync.SyncFrame)
	go conn.serve(ctx)
}

func (h *Handler) newChannelConn(c net.Conn, sync *eventing.SyncFrame) *channelConn {
	return &channelConn{
		rwc:     c,
		gateway: h.gateway,
		dec: encoding.NewDecoder(
			encoding.NewWorldPackReader(encoding.NewWorldReader(c, sync.Code)),
		),
		enc:      encoding.NewEncoder(encoding.NewWorldWriter(c)),
		Code:     sync.Code,
		sequence: sync.Sequence,
	}
}

type channelConn struct {
	rwc      net.Conn
	dec      *encoding.Decoder
	enc      *encoding.Encoder
	gateway  eventing.GatewayClient
	Code     uint32
	sequence uint32
}

func (c *channelConn) serve(ctx context.Context) {
	_, span := otel.Tracer(name).Start(ctx, "Handle Messages")
	defer span.End()
	logrus.Debugf("gateway: serving connection from %v", c.rwc.RemoteAddr())
	err := c.handleMessages(ctx)
	switch e := err.(type) {
	case *protonostale.Status:
		if err := c.enc.Encode(e); err != nil {
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

func (c *channelConn) handleMessages(ctx context.Context) error {
	stream, err := c.gateway.ChannelInteract(ctx)
	if err != nil {
		return protonostale.NewStatus(eventing.Code_UNEXPECTED_ERROR)
	}
	logrus.Debugf("gateway: created auth handoff interact stream")
	if err := stream.Send(&eventing.ChannelInteractRequest{
		Payload: &eventing.ChannelInteractRequest_SyncFrame{
			SyncFrame: &eventing.SyncFrame{
				Sequence: c.sequence,
				Code:     c.Code,
			},
		},
	}); err != nil {
		return protonostale.NewStatus(eventing.Code_UNEXPECTED_ERROR)
	}
	logrus.Debugf("gateway: sent sync frame")
	var identifier protonostale.IdentifierFrame
	if err := c.dec.Decode(&identifier); err != nil {
		return protonostale.NewStatus(eventing.Code_BAD_CASE)
	}
	logrus.Debugf("gateway: read identifier frame: %v", identifier.String())
	if err := stream.Send(&eventing.ChannelInteractRequest{
		Payload: &eventing.ChannelInteractRequest_IdentifierFrame{
			IdentifierFrame: &identifier.IdentifierFrame,
		},
	}); err != nil {
		return protonostale.NewStatus(eventing.Code_UNEXPECTED_ERROR)
	}
	logrus.Debugf("gateway: sent login frame")
	var password protonostale.PasswordFrame
	if err := c.dec.Decode(&password); err != nil {
		return protonostale.NewStatus(eventing.Code_BAD_CASE)
	}
	logrus.Debugf("gateway: read password frame: %v", password.String())
	if err := stream.Send(&eventing.ChannelInteractRequest{
		Payload: &eventing.ChannelInteractRequest_PasswordFrame{
			PasswordFrame: &password.PasswordFrame,
		},
	}); err != nil {
		return protonostale.NewStatus(eventing.Code_UNEXPECTED_ERROR)
	}
	logrus.Debugf("gateway: sent password frame")
	if err != nil {
		return protonostale.NewStatus(eventing.Code_UNEXPECTED_ERROR)
	}
	for {
		var m protonostale.ChannelInteractRequest
		if err := c.dec.Decode(&m); err != nil {
			return protonostale.NewStatus(eventing.Code_UNEXPECTED_ERROR)
		}
		if err := stream.Send(&m.ChannelInteractRequest); err != nil {
			return protonostale.NewStatus(eventing.Code_UNEXPECTED_ERROR)
		}
	}
}
