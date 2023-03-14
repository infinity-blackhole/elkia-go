package gateway

import (
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
	if err != nil {
		logrus.Errorf("auth: error while creating auth interact stream: %v", err)
		return
	}
	logrus.Debugf("gateway: created auth handoff interact stream")
	proxy, err := NewProxyUpgrader(c).Upgrade(stream)
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

func NewProxy(c net.Conn, code uint32) *Proxy {
	return &Proxy{
		sender: NewProxySender(c, code),
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
	conn  net.Conn
	proxy *SessionProxyClient
}

func NewProxyUpgrader(c net.Conn) *ProxyUpgrader {
	return &ProxyUpgrader{
		conn:  c,
		proxy: NewSessionProxyClient(c),
	}
}

func (p *ProxyUpgrader) Upgrade(
	stream eventing.Gateway_ChannelInteractClient,
) (*Proxy, error) {
	msg, err := p.proxy.Recv()
	if err != nil {
		return nil, err
	}
	logrus.Debugf("gateway: received sync frame: %v", msg.String())
	if err := stream.Send(msg); err != nil {
		return nil, p.proxy.SendMsg(
			protonostale.NewStatus(eventing.Code_UNEXPECTED_ERROR),
		)
	}
	return NewProxy(p.conn, msg.GetSyncFrame().GetCode()), nil
}

type SessionProxyClient struct {
	dec *SessionDecoder
	enc *Encoder
}

func NewSessionProxyClient(c net.Conn) *SessionProxyClient {
	return &SessionProxyClient{
		dec: NewSessionDecoder(c),
		enc: NewEncoder(c),
	}
}

func (c *SessionProxyClient) Recv() (*eventing.ChannelInteractRequest, error) {
	var msg protonostale.SyncFrame
	if err := c.RecvMsg(&msg); err != nil {
		return nil, c.SendMsg(
			protonostale.NewStatus(eventing.Code_BAD_CASE),
		)
	}
	return &eventing.ChannelInteractRequest{
		Payload: &eventing.ChannelInteractRequest_SyncFrame{
			SyncFrame: msg.SyncFrame,
		},
	}, nil
}

func (c *SessionProxyClient) RecvMsg(msg any) error {
	return c.dec.Decode(msg)
}

func (u *SessionProxyClient) Send(msg *eventing.ChannelInteractResponse) error {
	return u.enc.Encode(&protonostale.ChannelInteractResponse{
		ChannelInteractResponse: msg,
	})
}

func (c *SessionProxyClient) SendMsg(msg any) error {
	switch msg.(type) {
	case protonostale.Marshaler:
		return c.enc.Encode(msg)
	default:
		return c.enc.Encode(
			protonostale.NewStatus(eventing.Code_UNEXPECTED_ERROR),
		)
	}
}

type ProxySender struct {
	dec   *ChannelDecoder
	proxy *ChannelProxyClient
}

func NewProxySender(c net.Conn, code uint32) *ProxySender {
	return &ProxySender{
		dec:   NewChannelDecoder(c, code),
		proxy: NewChannelProxyClient(c, code),
	}
}

func (c *ProxySender) Serve(stream eventing.Gateway_ChannelInteractClient) error {
	if err := c.handleIdentifier(stream); err != nil {
		return err
	}
	logrus.Debugf("gateway: sent sync frame")
	if err := c.handlePassword(stream); err != nil {
		return err
	}
	logrus.Debugf("gateway: sent login frame")
	for {
		msg, err := c.proxy.Recv()
		if err != nil {
			return protonostale.NewStatus(eventing.Code_UNEXPECTED_ERROR)
		}
		if err := stream.Send(msg); err != nil {
			return protonostale.NewStatus(eventing.Code_UNEXPECTED_ERROR)
		}
	}
}

func (c *ProxySender) handleIdentifier(stream eventing.Gateway_ChannelInteractClient) error {
	var msg protonostale.IdentifierFrame
	if err := c.dec.Decode(&msg); err != nil {
		return protonostale.NewStatus(eventing.Code_BAD_CASE)
	}
	logrus.Debugf("gateway: read identifier frame: %v", msg.String())
	if err := stream.Send(&eventing.ChannelInteractRequest{
		Payload: &eventing.ChannelInteractRequest_IdentifierFrame{
			IdentifierFrame: msg.IdentifierFrame,
		},
	}); err != nil {
		return protonostale.NewStatus(eventing.Code_UNEXPECTED_ERROR)
	}
	return nil
}

func (c *ProxySender) handlePassword(stream eventing.Gateway_ChannelInteractClient) error {
	var msg protonostale.PasswordFrame
	if err := c.dec.Decode(&msg); err != nil {
		return err
	}
	logrus.Debugf("gateway: read password frame: %v", msg.String())
	if err := stream.Send(&eventing.ChannelInteractRequest{
		Payload: &eventing.ChannelInteractRequest_PasswordFrame{
			PasswordFrame: msg.PasswordFrame,
		},
	}); err != nil {
		return protonostale.NewStatus(eventing.Code_UNEXPECTED_ERROR)
	}
	return nil
}

type ChannelProxyClient struct {
	dec *ChannelDecoder
	enc *Encoder
}

func NewChannelProxyClient(c net.Conn, code uint32) *ChannelProxyClient {
	return &ChannelProxyClient{
		dec: NewChannelDecoder(c, code),
		enc: NewEncoder(c),
	}
}

func (c *ChannelProxyClient) Recv() (*eventing.ChannelInteractRequest, error) {
	var msg protonostale.ChannelInteractRequest
	if err := c.RecvMsg(&msg); err != nil {
		return nil, c.SendMsg(
			protonostale.NewStatus(eventing.Code_BAD_CASE),
		)
	}
	return msg.ChannelInteractRequest, nil
}

func (c *ChannelProxyClient) RecvMsg(msg any) error {
	return c.dec.Decode(msg)
}

func (u *ChannelProxyClient) Send(msg *eventing.ChannelInteractResponse) error {
	return u.enc.Encode(&protonostale.ChannelInteractResponse{
		ChannelInteractResponse: msg,
	})
}

func (c *ChannelProxyClient) SendMsg(msg any) error {
	switch msg.(type) {
	case protonostale.Marshaler:
		return c.enc.Encode(msg)
	default:
		return c.enc.Encode(
			protonostale.NewStatus(eventing.Code_UNEXPECTED_ERROR),
		)
	}
}
