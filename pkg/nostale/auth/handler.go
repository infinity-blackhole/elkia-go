package auth

import (
	"context"
	"net"

	eventing "github.com/infinity-blackhole/elkia/pkg/api/eventing/v1alpha1"
	"github.com/infinity-blackhole/elkia/pkg/protonostale"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
)

const name = "github.com/infinity-blackhole/elkia/internal/auth"

type HandlerConfig struct {
	AuthClient eventing.AuthClient
}

func NewHandler(cfg HandlerConfig) *Handler {
	return &Handler{
		auth: cfg.AuthClient,
	}
}

type Handler struct {
	auth eventing.AuthClient
}

func (h *Handler) ServeNosTale(c net.Conn) {
	ctx := context.Background()
	conn := h.newConn(c)
	logrus.Debugf("auth: new connection from %v", c.RemoteAddr())
	go conn.serve(ctx)
}

func (h *Handler) newConn(c net.Conn) *conn {
	return &conn{
		rwc:  c,
		dec:  NewDecoder(NewLoginEncoding(), c),
		enc:  NewEncoder(NewLoginEncoding(), c),
		auth: h.auth,
	}
}

type conn struct {
	rwc  net.Conn
	dec  *Decoder
	enc  *Encoder
	auth eventing.AuthClient
}

func (c *conn) serve(ctx context.Context) {
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

func (c *conn) handleMessages(ctx context.Context) error {
	for {
		var msg protonostale.AuthInteractRequest
		if err := c.dec.Decode(&msg); err != nil {
			return protonostale.NewStatus(eventing.DialogErrorCode_UNEXPECTED_ERROR)
		}
		logrus.Debugf("auth: read event: %v", msg.Payload)
		stream, err := c.auth.AuthInteract(ctx)
		if err != nil {
			return protonostale.NewStatus(eventing.DialogErrorCode_UNEXPECTED_ERROR)
		}
		logrus.Debugf("auth: created auth interact stream")
		if err := stream.Send(&msg.AuthInteractRequest); err != nil {
			return protonostale.NewStatus(eventing.DialogErrorCode_UNEXPECTED_ERROR)
		}
		logrus.Debug("auth: sent login request")
		m, err := stream.Recv()
		if err != nil {
			return protonostale.NewStatus(eventing.DialogErrorCode_UNEXPECTED_ERROR)
		}
		logrus.Debugf("auth: received login response: %v", m)
		switch m.Payload.(type) {
		case *eventing.AuthInteractResponse_EndpointListEvent:
			ed := protonostale.EndpointListEvent{
				EndpointListEvent: *m.GetEndpointListEvent(),
			}
			if err := c.enc.Encode(&ed); err != nil {
				return protonostale.NewStatus(eventing.DialogErrorCode_UNEXPECTED_ERROR)
			}
			logrus.Debug("auth: wrote endpoint list event")
		default:
			logrus.Errorf("auth: unexpected login response: %v", m)
			return protonostale.NewStatus(eventing.DialogErrorCode_BAD_CASE)
		}
	}
}
