package auth

import (
	"bufio"
	"context"
	"net"

	eventing "github.com/infinity-blackhole/elkia/pkg/api/eventing/v1alpha1"
	"github.com/infinity-blackhole/elkia/pkg/protonostale"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
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
		rc:   bufio.NewReader(NewReader(bufio.NewReader(c))),
		wc:   bufio.NewWriter(NewWriter(bufio.NewWriter(c))),
		auth: h.auth,
	}
}

type conn struct {
	rwc  net.Conn
	rc   *bufio.Reader
	wc   *bufio.Writer
	auth eventing.AuthClient
}

func (c *conn) serve(ctx context.Context) {
	go func() {
		if err := recover(); err != nil {
			if _, err := protonostale.WriteDialogErrorEvent(
				c.wc,
				&eventing.DialogErrorEvent{
					Code: eventing.DialogErrorCode_UNEXPECTED_ERROR,
				},
			); err != nil {
				logrus.Fatal(err)
			}
		}
	}()
	_, span := otel.Tracer(name).Start(ctx, "Serve")
	defer span.End()
	scanner := bufio.NewScanner(c.rc)
	scanner.Split(bufio.ScanLines)
	if !scanner.Scan() {
		if _, err := protonostale.WriteDialogErrorEvent(
			c.wc,
			&eventing.DialogErrorEvent{
				Code: eventing.DialogErrorCode_BAD_CASE,
			},
		); err != nil {
			logrus.Fatal(err)
		}
		if err := scanner.Err(); err != nil {
			logrus.Debugf("auth: read opcode: %v", err)
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
		return
	}
	event, err := protonostale.ParseAuthEvent(scanner.Text())
	if err != nil {
		if _, err := protonostale.WriteDialogErrorEvent(
			c.wc,
			&eventing.DialogErrorEvent{
				Code: eventing.DialogErrorCode_BAD_CASE,
			},
		); err != nil {
			logrus.Fatal(err)
		}
		logrus.Debugf("auth: read opcode: %v", err)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return
	}
	logrus.Debugf("auth: read opcode: %v", event)
	stream, err := c.auth.AuthInteract(ctx)
	if err != nil {
		logrus.Fatal(err)
	}
	if err := stream.Send(event); err != nil {
		logrus.Debugf("auth: send event: %v", err)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return
	}
	m, err := stream.Recv()
	if err != nil {
		logrus.Debugf("auth: recv event: %v", err)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return
	}
	switch m.Payload.(type) {
	case *eventing.AuthInteractResponse_EndpointListEvent:
		if _, err := protonostale.WriteEndpointListEvent(
			c.wc,
			m.GetEndpointListEvent(),
		); err != nil {
			logrus.Fatal(err)
		}
	}
}
