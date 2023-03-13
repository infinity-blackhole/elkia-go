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
	conn.serve(ctx)
}

func (h *Handler) newConn(c net.Conn) *conn {
	return &conn{
		rwc:  c,
		dec:  NewDecoder(c),
		enc:  NewEncoder(c),
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
	switch err := c.handleFrames(ctx).(type) {
	case protonostale.Marshaler:
		if err := c.enc.Encode(err); err != nil {
			logrus.Errorf("auth: failed to send error: %v", err)
		}
	default:
		if err := c.enc.Encode(
			protonostale.NewStatus(eventing.Code_UNEXPECTED_ERROR),
		); err != nil {
			logrus.Errorf("auth: failed to send error: %v", err)
		}
	}
}

func (c *conn) handleFrames(ctx context.Context) error {
	stream, err := c.auth.AuthInteract(ctx)
	if err != nil {
		return protonostale.NewStatus(eventing.Code_UNEXPECTED_ERROR)
	}
	logrus.Debugf("auth: created auth interact stream")
	for {
		var msg protonostale.AuthInteractRequest
		if err := c.dec.Decode(&msg); err != nil {
			return protonostale.NewStatus(eventing.Code_UNEXPECTED_ERROR)
		}
		logrus.Debugf("auth: read frame: %v", msg.Payload)
		if err := stream.Send(&msg.AuthInteractRequest); err != nil {
			return protonostale.NewStatus(eventing.Code_UNEXPECTED_ERROR)
		}
		logrus.Debug("auth: sent login request")
		m, err := stream.Recv()
		if err != nil {
			return protonostale.NewStatus(eventing.Code_UNEXPECTED_ERROR)
		}
		logrus.Debugf("auth: received login response: %v", m)
		switch p := m.Payload.(type) {
		case *eventing.AuthInteractResponse_EndpointListFrame:
			ed := protonostale.EndpointListFrame{
				EndpointListFrame: eventing.EndpointListFrame{
					Code:      p.EndpointListFrame.Code,
					Endpoints: p.EndpointListFrame.Endpoints,
				},
			}
			if err := c.enc.Encode(&ed); err != nil {
				return protonostale.NewStatus(eventing.Code_UNEXPECTED_ERROR)
			}
			logrus.Debug("auth: wrote endpoint list frame")
		default:
			logrus.Errorf("auth: unexpected login response: %v", m)
			return protonostale.NewStatus(eventing.Code_BAD_CASE)
		}
	}
}
