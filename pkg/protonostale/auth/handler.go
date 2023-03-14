package auth

import (
	"context"
	"net"

	eventing "github.com/infinity-blackhole/elkia/pkg/api/eventing/v1alpha1"
	"github.com/infinity-blackhole/elkia/pkg/protonostale"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"golang.org/x/sync/errgroup"
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
	_, span := otel.Tracer(name).Start(ctx, "Handle Messages")
	defer span.End()
	sender := h.newProxySender(ctx, c)
	receiver := h.newProxyReceiver(ctx, c)
	logrus.Debugf("auth: new connection from %v", c.RemoteAddr())
	stream, err := h.auth.AuthInteract(ctx)
	if err != nil {
		logrus.Errorf("auth: error while creating auth interact stream: %v", err)
		return
	}
	logrus.Debugf("auth: created auth interact stream")
	ws := errgroup.Group{}
	ws.Go(func() error {
		return sender.Serve(stream)
	})
	ws.Go(func() error {
		return receiver.Serve(stream)
	})
	if err := ws.Wait(); err != nil {
		logrus.Errorf("auth: error while handling connection: %v", err)
	}
}

func (h *Handler) newProxySender(ctx context.Context, c net.Conn) *ProxySender {
	return &ProxySender{
		ctx:   ctx,
		proxy: NewProxyClient(ctx, c),
	}
}

func (h *Handler) newProxyReceiver(ctx context.Context, c net.Conn) *ProxyReceiver {
	return &ProxyReceiver{
		ctx:   ctx,
		proxy: NewProxyClient(ctx, c),
	}
}

type ProxySender struct {
	ctx   context.Context
	proxy *ProxyClient
}

func (p *ProxySender) Serve(stream eventing.Auth_AuthInteractClient) error {
	for {
		msg, err := p.proxy.Recv()
		if err != nil {
			return p.proxy.SendMsg(
				protonostale.NewStatus(eventing.Code_UNEXPECTED_ERROR),
			)
		}
		logrus.Debugf("auth: read frame: %v", msg.Payload)
		if err := stream.Send(msg); err != nil {
			return p.proxy.SendMsg(
				protonostale.NewStatus(eventing.Code_UNEXPECTED_ERROR),
			)
		}
		logrus.Debug("auth: sent login request")
	}
}

type ProxyReceiver struct {
	ctx   context.Context
	proxy *ProxyClient
}

func (c *ProxyReceiver) Serve(stream eventing.Auth_AuthInteractClient) error {
	for {
		msg, err := stream.Recv()
		if err != nil {
			return c.proxy.SendMsg(
				protonostale.NewStatus(eventing.Code_UNEXPECTED_ERROR),
			)
		}
		logrus.Debugf("auth: read frame: %v", msg.Payload)
		if err := c.proxy.Send(msg); err != nil {
			return c.proxy.SendMsg(
				protonostale.NewStatus(eventing.Code_UNEXPECTED_ERROR),
			)
		}
		logrus.Debug("auth: sent login request")
	}
}

type ProxyClient struct {
	ctx  context.Context
	conn net.Conn
	dec  *Decoder
	enc  *Encoder
}

func NewProxyClient(ctx context.Context, c net.Conn) *ProxyClient {
	return &ProxyClient{
		ctx:  ctx,
		conn: c,
		dec:  NewDecoder(c),
		enc:  NewEncoder(c),
	}
}

func (c *ProxyClient) Recv() (*eventing.AuthInteractRequest, error) {
	var msg protonostale.AuthInteractRequest
	if err := c.RecvMsg(&msg); err != nil {
		return nil, c.SendMsg(err)
	}
	return msg.AuthInteractRequest, nil
}

func (c *ProxyClient) RecvMsg(msg any) error {
	return c.dec.Decode(msg)
}

func (c *ProxyClient) Send(msg *eventing.AuthInteractResponse) error {
	return c.SendMsg(&protonostale.AuthInteractResponse{
		AuthInteractResponse: msg,
	})
}

func (c *ProxyClient) SendMsg(msg any) error {
	switch msg.(type) {
	case protonostale.Marshaler:
		return c.enc.Encode(msg)
	default:
		return c.enc.Encode(
			protonostale.NewStatus(eventing.Code_UNEXPECTED_ERROR),
		)
	}
}
