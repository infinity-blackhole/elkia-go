package auth

import (
	"bufio"
	"context"
	"fmt"
	"net"

	eventing "github.com/infinity-blackhole/elkia/pkg/api/eventing/v1alpha1"
	"github.com/infinity-blackhole/elkia/pkg/nostale/utils"
	"github.com/infinity-blackhole/elkia/pkg/protonostale"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
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
			utils.WriteError(
				c.wc,
				eventing.DialogErrorCode_UNEXPECTED_ERROR,
				"An unexpected error occurred, please try again later",
			)
		}
	}()
	logrus.Debugf("auth: serving connection from %v", c.rwc.RemoteAddr())
	c.handleMessages(ctx)
}

func (c *conn) handleMessages(ctx context.Context) {
	_, span := otel.Tracer(name).Start(ctx, "Handle Messages")
	defer span.End()
	scanner := bufio.NewScanner(c.rc)
	scanner.Split(bufio.ScanLines)
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			utils.WriteError(
				c.wc,
				eventing.DialogErrorCode_BAD_CASE,
				fmt.Sprintf(
					"failed to read auth event: %v",
					err,
				),
			)
			return
		}
		utils.WriteError(
			c.wc,
			eventing.DialogErrorCode_BAD_CASE,
			"failed to read auth event: EOF",
		)
		return
	}
	logrus.Debugf("auth: read message: %v", scanner.Text())
	event, err := protonostale.ParseAuthEvent(scanner.Text())
	if err != nil {
		utils.WriteError(
			c.wc,
			eventing.DialogErrorCode_BAD_CASE,
			fmt.Sprintf(
				"failed to parse auth event: %v",
				err,
			),
		)
		return
	}
	logrus.Debugf("auth: read event: %v", event)
	c.handleAuthLogin(ctx, event)
}

func (c *conn) handleAuthLogin(
	ctx context.Context,
	event *eventing.AuthInteractRequest,
) {
	span := trace.SpanFromContext(ctx)
	defer span.End()
	stream, err := c.auth.AuthInteract(ctx)
	if err != nil {
		utils.WriteError(
			c.wc,
			eventing.DialogErrorCode_UNEXPECTED_ERROR,
			fmt.Sprintf(
				"failed to create auth interact stream: %v",
				err,
			),
		)
		return
	}
	logrus.Debugf("auth: created auth interact stream")
	if err := stream.Send(event); err != nil {
		utils.WriteError(
			c.wc,
			eventing.DialogErrorCode_UNEXPECTED_ERROR,
			fmt.Sprintf(
				"failed to send login request: %v",
				err,
			),
		)
		return
	}
	logrus.Debug("auth: sent login request")
	m, err := stream.Recv()
	if err != nil {
		utils.WriteError(
			c.wc,
			eventing.DialogErrorCode_UNEXPECTED_ERROR,
			fmt.Sprintf(
				"failed to receive login response: %v",
				err,
			),
		)
		return
	}
	logrus.Debugf("auth: received login response: %v", m)
	switch m.Payload.(type) {
	case *eventing.AuthInteractResponse_EndpointListEvent:
		if _, err := protonostale.WriteEndpointListEvent(
			c.wc,
			m.GetEndpointListEvent(),
		); err != nil {
			utils.WriteError(
				c.wc,
				eventing.DialogErrorCode_UNEXPECTED_ERROR,
				fmt.Sprintf(
					"failed to write endpoint list event: %v",
					err,
				),
			)
			return
		}
		logrus.Debug("auth: wrote endpoint list event")
	default:
		utils.WriteError(
			c.wc,
			eventing.DialogErrorCode_UNEXPECTED_ERROR,
			fmt.Sprintf(
				"unexpected login response: %v",
				m,
			),
		)
		logrus.Errorf("auth: unexpected login response: %v", m)
	}
}
