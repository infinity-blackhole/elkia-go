package auth

import (
	"bufio"
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
	logrus.Debugf("auth: new connection from %v", c.RemoteAddr())
	stream, err := h.auth.AuthInteract(ctx)
	if err != nil {
		logrus.Errorf("auth: error while creating auth interact stream: %v", err)
		return
	}
	logrus.Debugf("auth: created auth interact stream")
	if err := NewProxy(bufio.NewReadWriter(
		bufio.NewReader(c),
		bufio.NewWriter(c),
	)).Serve(stream); err != nil {
		logrus.Errorf("auth: error while handling connection: %v", err)
	}
}

type Proxy struct {
	s *ProxySender
	r *ProxyReceiver
}

func NewProxy(rw *bufio.ReadWriter) *Proxy {
	return &Proxy{
		s: NewProxySender(rw),
		r: NewProxyReceiver(rw),
	}
}

func (p *Proxy) Serve(stream eventing.Auth_AuthInteractClient) error {
	ws := errgroup.Group{}
	ws.Go(func() error {
		return p.s.Serve(stream)
	})
	ws.Go(func() error {
		return p.r.Serve(stream)
	})
	return ws.Wait()
}

type ProxySender struct {
	proxy *ProxyClient
}

func NewProxySender(rw *bufio.ReadWriter) *ProxySender {
	return &ProxySender{
		proxy: NewProxyClient(rw),
	}
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
	proxy *ProxyClient
}

func NewProxyReceiver(rw *bufio.ReadWriter) *ProxyReceiver {
	return &ProxyReceiver{
		proxy: NewProxyClient(rw),
	}
}

func (p *ProxyReceiver) Serve(stream eventing.Auth_AuthInteractClient) error {
	for {
		msg, err := stream.Recv()
		if err != nil {
			return p.proxy.SendMsg(
				protonostale.NewStatus(eventing.Code_UNEXPECTED_ERROR),
			)
		}
		logrus.Debugf("auth: read frame: %v", msg.Payload)
		if err := p.proxy.Send(msg); err != nil {
			return p.proxy.SendMsg(
				protonostale.NewStatus(eventing.Code_UNEXPECTED_ERROR),
			)
		}
		logrus.Debug("auth: sent login request")
	}
}

type ProxyClient struct {
	rw  *bufio.ReadWriter
	dec *Decoder
	enc *Encoder
}

func NewProxyClient(rw *bufio.ReadWriter) *ProxyClient {
	return &ProxyClient{
		rw:  rw,
		dec: NewDecoder(rw),
		enc: NewEncoder(rw),
	}
}

func (p *ProxyClient) Recv() (*eventing.AuthInteractRequest, error) {
	var msg protonostale.AuthInteractRequest
	if err := p.RecvMsg(&msg); err != nil {
		return nil, p.SendMsg(
			protonostale.NewStatus(eventing.Code_BAD_CASE),
		)
	}
	return msg.AuthInteractRequest, nil
}

func (p *ProxyClient) RecvMsg(msg any) error {
	return p.dec.Decode(msg)
}

func (p *ProxyClient) Send(msg *eventing.AuthInteractResponse) error {
	return p.SendMsg(&protonostale.AuthInteractResponse{
		AuthInteractResponse: msg,
	})
}

func (p *ProxyClient) SendMsg(msg any) error {
	switch msg.(type) {
	case protonostale.Marshaler:
		if err := p.enc.Encode(msg); err != nil {
			return err
		}
	default:
		if err := p.enc.Encode(
			protonostale.NewStatus(eventing.Code_UNEXPECTED_ERROR),
		); err != nil {
			return err
		}
	}
	return p.rw.Flush()
}
