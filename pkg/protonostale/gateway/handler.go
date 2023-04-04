package gateway

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
	_, span := otel.Tracer(name).Start(ctx, "Handle Messages")
	defer span.End()
	stream, err := h.gateway.ChannelInteract(ctx)
	defer stream.CloseSend()
	if err != nil {
		logrus.Errorf("auth: error while creating auth interact stream: %v", err)
		return
	}
	logrus.Debugf("gateway: created auth handoff interact stream")
	proxy, err := NewProxyUpgrader(bufio.NewReadWriter(
		bufio.NewReader(c),
		bufio.NewWriter(c),
	)).Upgrade(stream)
	if err != nil {
		logrus.Errorf("gateway: error while upgrading connection: %v", err)
		return
	}
	if err := proxy.Serve(stream); err != nil {
		logrus.Errorf("auth: error while handling connection: %v", err)
	}
}

type Proxy struct {
	sender *ProxySender
}

func NewProxy(rw *bufio.ReadWriter, code uint32) *Proxy {
	return &Proxy{
		sender: NewProxySender(rw, code),
	}
}

func (p *Proxy) Serve(stream eventing.Gateway_ChannelInteractClient) error {
	ws := errgroup.Group{}
	ws.Go(func() error {
		return p.sender.Serve(stream)
	})
	return ws.Wait()
}

type ProxyUpgrader struct {
	rwc   *bufio.ReadWriter
	proxy *SessionProxyClient
}

func NewProxyUpgrader(rwc *bufio.ReadWriter) *ProxyUpgrader {
	return &ProxyUpgrader{
		rwc:   rwc,
		proxy: NewSessionProxyClient(rwc),
	}
}

func (p *ProxyUpgrader) Upgrade(
	stream eventing.Gateway_ChannelInteractClient,
) (*Proxy, error) {
	msg, err := p.proxy.RecvSync()
	if err != nil {
		return nil, err
	}
	logrus.Debugf("gateway: received sync frame: %v", msg.String())
	if err := stream.Send(msg); err != nil {
		return nil, p.proxy.SendMsg(
			protonostale.NewStatus(eventing.Code_UNEXPECTED_ERROR),
		)
	}
	return NewProxy(p.rwc, 0), nil
}

type SessionProxyClient struct {
	rw  *bufio.ReadWriter
	dec *SessionDecoder
	enc *Encoder
}

func NewSessionProxyClient(rw *bufio.ReadWriter) *SessionProxyClient {
	return &SessionProxyClient{
		rw:  rw,
		dec: NewSessionDecoder(rw.Reader),
		enc: NewEncoder(rw),
	}
}

func (p *SessionProxyClient) RecvSync() (*eventing.ChannelInteractRequest, error) {
	var msg protonostale.SyncFrame
	if err := p.RecvMsg(&msg); err != nil {
		if err := p.SendMsg(
			protonostale.NewStatus(eventing.Code_BAD_CASE),
		); err != nil {
			return nil, err
		}
		return nil, err
	}
	return &eventing.ChannelInteractRequest{
		Payload: &eventing.ChannelInteractRequest_SyncFrame{
			SyncFrame: msg.SyncFrame,
		},
	}, nil
}

func (p *SessionProxyClient) RecvMsg(msg any) error {
	return p.dec.Decode(msg)
}

func (u *SessionProxyClient) Send(msg *eventing.ChannelInteractResponse) error {
	return u.enc.Encode(&protonostale.ChannelInteractResponse{
		ChannelInteractResponse: msg,
	})
}

func (p *SessionProxyClient) SendMsg(msg any) error {
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

type ProxySender struct {
	proxy *ChannelProxyClient
}

func NewProxySender(rw *bufio.ReadWriter, code uint32) *ProxySender {
	return &ProxySender{
		proxy: NewChannelProxyClient(rw, code),
	}
}

func (p *ProxySender) Serve(stream eventing.Gateway_ChannelInteractClient) error {
	msg, err := p.proxy.RecvIdentifier()
	if err != nil {
		return protonostale.NewStatus(eventing.Code_UNEXPECTED_ERROR)
	}
	if err := stream.Send(msg); err != nil {
		return protonostale.NewStatus(eventing.Code_UNEXPECTED_ERROR)
	}
	logrus.Debugf("gateway: sent sync frame")
	msg, err = p.proxy.RecvPassword()
	if err != nil {
		return protonostale.NewStatus(eventing.Code_UNEXPECTED_ERROR)
	}
	if err := stream.Send(msg); err != nil {
		return protonostale.NewStatus(eventing.Code_UNEXPECTED_ERROR)
	}
	logrus.Debugf("gateway: sent login frame")
	for {
		msg, err = p.proxy.RecvCommand()
		if err != nil {
			return protonostale.NewStatus(eventing.Code_UNEXPECTED_ERROR)
		}
		if err := stream.Send(msg); err != nil {
			return protonostale.NewStatus(eventing.Code_UNEXPECTED_ERROR)
		}
	}
}

type ChannelProxyClient struct {
	rw  *bufio.ReadWriter
	dec *ChannelDecoder
	enc *Encoder
}

func NewChannelProxyClient(rw *bufio.ReadWriter, code uint32) *ChannelProxyClient {
	return &ChannelProxyClient{
		rw:  rw,
		dec: NewChannelDecoder(rw, code),
		enc: NewEncoder(rw),
	}
}

func (p *ChannelProxyClient) RecvIdentifier() (*eventing.ChannelInteractRequest, error) {
	var msg protonostale.IdentifierFrame
	if err := p.RecvMsg(&msg); err != nil {
		if err := p.SendMsg(
			protonostale.NewStatus(eventing.Code_BAD_CASE),
		); err != nil {
			return nil, err
		}
		return nil, err
	}
	return &eventing.ChannelInteractRequest{
		Payload: &eventing.ChannelInteractRequest_IdentifierFrame{
			IdentifierFrame: msg.IdentifierFrame,
		},
	}, nil
}

func (p *ChannelProxyClient) RecvPassword() (*eventing.ChannelInteractRequest, error) {
	var msg protonostale.PasswordFrame
	if err := p.RecvMsg(&msg); err != nil {
		if err := p.SendMsg(
			protonostale.NewStatus(eventing.Code_BAD_CASE),
		); err != nil {
			return nil, err
		}
		return nil, err
	}
	return &eventing.ChannelInteractRequest{
		Payload: &eventing.ChannelInteractRequest_PasswordFrame{
			PasswordFrame: msg.PasswordFrame,
		},
	}, nil
}

func (p *ChannelProxyClient) RecvCommand() (*eventing.ChannelInteractRequest, error) {
	var msg protonostale.CommandFrame
	if err := p.RecvMsg(&msg); err != nil {
		if err := p.SendMsg(
			protonostale.NewStatus(eventing.Code_BAD_CASE),
		); err != nil {
			return nil, err
		}
		return nil, err
	}
	return &eventing.ChannelInteractRequest{
		Payload: &eventing.ChannelInteractRequest_CommandFrame{
			CommandFrame: msg.CommandFrame,
		},
	}, nil
}

func (p *ChannelProxyClient) RecvMsg(msg any) error {
	return p.dec.Decode(msg)
}

func (u *ChannelProxyClient) Send(msg *eventing.ChannelInteractResponse) error {
	return u.enc.Encode(&protonostale.ChannelInteractResponse{
		ChannelInteractResponse: msg,
	})
}

func (p *ChannelProxyClient) SendMsg(msg any) error {
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
