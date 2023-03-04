package gateway

import (
	"context"
	"net"

	eventing "github.com/infinity-blackhole/elkia/pkg/api/eventing/v1alpha1"
	"github.com/infinity-blackhole/elkia/pkg/nostale/encoding"
	"github.com/infinity-blackhole/elkia/pkg/nostale/utils"
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
	conn := h.newHandoffConn(c)
	logrus.Debugf("gateway: new handshaker from %v", c.RemoteAddr())
	go conn.serve(ctx)
}

func (h *Handler) newHandoffConn(c net.Conn) *handoffConn {
	return &handoffConn{
		rwc:     c,
		dec:     encoding.NewDecoder(utils.NewReadLogger("session: ", c), encoding.SessionEncoding),
		gateway: h.gateway,
	}
}

type handoffConn struct {
	rwc     net.Conn
	dec     *encoding.Decoder
	gateway eventing.GatewayClient
}

func (c *handoffConn) serve(ctx context.Context) {
	_, span := otel.Tracer(name).Start(ctx, "Handle Messages")
	defer span.End()
	logrus.Debugf("gateway: serving connection from %v", c.rwc.RemoteAddr())
	if err := c.handleMessages(ctx); err != nil {
		c.rwc.Close()
	}
}

func (c *handoffConn) handleMessages(ctx context.Context) error {
	var sync protonostale.SyncFrame
	if err := c.dec.Decode(&sync); err != nil {
		return protonostale.NewStatus(eventing.Code_UNEXPECTED_ERROR)
	}
	logrus.Debugf("gateway: received sync frame: %v", sync.String())
	conn := c.newChannelConn(&sync.SyncFrame)
	go conn.serve(ctx)
	return nil
}

func (h *handoffConn) newChannelConn(sync *eventing.SyncFrame) *channelConn {
	e := encoding.WorldEncoding.WithKey(sync.Code)
	logrus.Debugf("gateway: using encoding: %v", e)
	return &channelConn{
		rwc:      h.rwc,
		gateway:  h.gateway,
		dec:      encoding.NewDecoder(utils.NewReadLogger("gateway: ", h.rwc), e),
		enc:      encoding.NewEncoder(h.rwc, e),
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
	var handoffLogin protonostale.HandoffFrame
	if err := c.dec.Decode(&handoffLogin); err != nil {
		return protonostale.NewStatus(eventing.Code_BAD_CASE)
	}
	logrus.Debugf("gateway: read frame: %v", handoffLogin.String())
	if err := stream.Send(&eventing.ChannelInteractRequest{
		Payload: &eventing.ChannelInteractRequest_HandoffFrame{
			HandoffFrame: &handoffLogin.HandoffFrame,
		},
	}); err != nil {
		return protonostale.NewStatus(eventing.Code_UNEXPECTED_ERROR)
	}
	logrus.Debugf("gateway: sent login frame")
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
