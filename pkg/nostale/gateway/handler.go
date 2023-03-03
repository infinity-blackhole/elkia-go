package gateway

import (
	"context"
	"net"
	"strconv"

	eventing "github.com/infinity-blackhole/elkia/pkg/api/eventing/v1alpha1"
	"github.com/infinity-blackhole/elkia/pkg/nostale/encoding"
	"github.com/infinity-blackhole/elkia/pkg/protonostale"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"google.golang.org/grpc/metadata"
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
		dec:     encoding.NewDecoder(c, encoding.SessionEncoding),
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
	logrus.Debugf("auth: serving connection from %v", c.rwc.RemoteAddr())
	if err := c.handleMessages(ctx); err != nil {
		c.rwc.Close()
	}
}

func (c *handoffConn) handleMessages(ctx context.Context) error {
	var sync protonostale.AuthHandoffSyncFrame
	if err := c.dec.Decode(&sync); err != nil {
		return protonostale.NewStatus(eventing.DialogErrorCode_UNEXPECTED_ERROR)
	}
	conn := c.newChannelConn(&sync.AuthHandoffSyncFrame)
	go conn.serve(ctx)
	return nil
}

func (h *handoffConn) newChannelConn(sync *eventing.AuthHandoffSyncFrame) *channelConn {
	e := encoding.WorldEncoding.WithKey(sync.Code)
	return &channelConn{
		rwc:      h.rwc,
		gateway:  h.gateway,
		dec:      encoding.NewDecoder(h.rwc, e),
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
	logrus.Debugf("auth: serving connection from %v", c.rwc.RemoteAddr())
	err := c.handleMessages(ctx)
	switch e := err.(type) {
	case *protonostale.Status:
		if err := c.enc.Encode(e); err != nil {
			logrus.Errorf("auth: failed to send error: %v", err)
		}
	default:
		if err := c.enc.Encode(
			protonostale.NewStatus(eventing.DialogErrorCode_UNEXPECTED_ERROR),
		); err != nil {
			logrus.Errorf("auth: failed to send error: %v", err)
		}
	}
}

func (c *channelConn) handleMessages(ctx context.Context) error {
	authStream, err := c.gateway.AuthHandoffInteract(ctx)
	if err != nil {
		return protonostale.NewStatus(eventing.DialogErrorCode_UNEXPECTED_ERROR)
	}
	logrus.Debugf("gateway: created auth handoff interact stream")
	if err := authStream.Send(&eventing.AuthHandoffInteractRequest{
		Payload: &eventing.AuthHandoffInteractRequest_SyncFrame{
			SyncFrame: &eventing.AuthHandoffSyncFrame{
				Sequence: c.sequence,
				Code:     c.Code,
			},
		},
	}); err != nil {
		return protonostale.NewStatus(eventing.DialogErrorCode_UNEXPECTED_ERROR)
	}
	var handoffLogin protonostale.AuthHandoffLoginFrame
	if err := c.dec.Decode(&handoffLogin); err != nil {
		return protonostale.NewStatus(eventing.DialogErrorCode_BAD_CASE)
	}
	logrus.Debugf("gateway: read frame: %v", handoffLogin)
	if err := authStream.Send(&eventing.AuthHandoffInteractRequest{
		Payload: &eventing.AuthHandoffInteractRequest_LoginFrame{
			LoginFrame: &handoffLogin.AuthHandoffLoginFrame,
		},
	}); err != nil {
		return protonostale.NewStatus(eventing.DialogErrorCode_UNEXPECTED_ERROR)
	}
	logrus.Debugf("gateway: sent login frame")
	m, err := authStream.Recv()
	if err != nil {
		return protonostale.NewStatus(eventing.DialogErrorCode_UNEXPECTED_ERROR)
	}
	logrus.Debugf("gateway: received login success frame")
	loginSuccess := m.GetLoginSuccessFrame()
	if loginSuccess == nil {
		return protonostale.NewStatus(eventing.DialogErrorCode_UNEXPECTED_ERROR)
	}
	logrus.Debugf("gateway: received login success frame: %v", loginSuccess)
	ctx = metadata.AppendToOutgoingContext(
		ctx,
		"sequence", strconv.FormatUint(uint64(c.sequence), 10),
		"code", strconv.FormatUint(uint64(c.Code), 10),
		"session", loginSuccess.Token,
	)
	channelStream, err := c.gateway.ChannelInteract(ctx)
	if err != nil {
		return protonostale.NewStatus(eventing.DialogErrorCode_UNEXPECTED_ERROR)
	}
	for {
		var m protonostale.ChannelFrame
		if err := c.dec.Decode(&m); err != nil {
			return protonostale.NewStatus(eventing.DialogErrorCode_UNEXPECTED_ERROR)
		}
		if err := channelStream.Send(&eventing.ChannelInteractRequest{
			Payload: &eventing.ChannelInteractRequest_ChannelFrame{
				ChannelFrame: &m.ChannelFrame,
			},
		}); err != nil {
			return protonostale.NewStatus(eventing.DialogErrorCode_UNEXPECTED_ERROR)
		}
	}
}
