package gateway

import (
	"bufio"
	"context"
	"io"
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

func NewProxy(rw io.ReadWriter, code uint32) *Proxy {
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
	return NewProxy(p.rwc, msg.GetSyncFrame().GetCode()), nil
}

type SessionProxyClient struct {
	dec *SessionDecoder
	enc *Encoder
}

func NewSessionProxyClient(rw *bufio.ReadWriter) *SessionProxyClient {
	return &SessionProxyClient{
		dec: NewSessionDecoder(rw.Reader),
		enc: NewEncoder(rw),
	}
}

func (p *SessionProxyClient) Recv() (*eventing.ChannelInteractRequest, error) {
	var msg protonostale.SyncFrame
	if err := p.RecvMsg(&msg); err != nil {
		return nil, p.SendMsg(
			protonostale.NewStatus(eventing.Code_BAD_CASE),
		)
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
		return p.enc.Encode(msg)
	default:
		return p.enc.Encode(
			protonostale.NewStatus(eventing.Code_UNEXPECTED_ERROR),
		)
	}
}

type ProxySender struct {
	dec   *ChannelDecoder
	proxy *ChannelProxyClient
}

func NewProxySender(rw io.ReadWriter, code uint32) *ProxySender {
	return &ProxySender{
		dec:   NewChannelDecoder(rw, code),
		proxy: NewChannelProxyClient(rw, code),
	}
}

func (p *ProxySender) Serve(stream eventing.Gateway_ChannelInteractClient) error {
	if err := p.handleIdentifier(stream); err != nil {
		return err
	}
	logrus.Debugf("gateway: sent sync frame")
	if err := p.handlePassword(stream); err != nil {
		return err
	}
	logrus.Debugf("gateway: sent login frame")
	for {
		msg, err := p.proxy.Recv()
		if err != nil {
			return protonostale.NewStatus(eventing.Code_UNEXPECTED_ERROR)
		}
		if err := stream.Send(msg); err != nil {
			return protonostale.NewStatus(eventing.Code_UNEXPECTED_ERROR)
		}
	}
}

func (p *ProxySender) handleIdentifier(stream eventing.Gateway_ChannelInteractClient) error {
	var msg protonostale.IdentifierFrame
	if err := p.dec.Decode(&msg); err != nil {
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

func (p *ProxySender) handlePassword(stream eventing.Gateway_ChannelInteractClient) error {
	var msg protonostale.PasswordFrame
	if err := p.dec.Decode(&msg); err != nil {
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

func NewChannelProxyClient(rw io.ReadWriter, code uint32) *ChannelProxyClient {
	return &ChannelProxyClient{
		dec: NewChannelDecoder(rw, code),
		enc: NewEncoder(rw),
	}
}

func (p *ChannelProxyClient) Recv() (*eventing.ChannelInteractRequest, error) {
	var msg protonostale.ChannelInteractRequest
	if err := p.RecvMsg(&msg); err != nil {
		return nil, p.SendMsg(
			protonostale.NewStatus(eventing.Code_BAD_CASE),
		)
	}
	return msg.ChannelInteractRequest, nil
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
		return p.enc.Encode(msg)
	default:
		return p.enc.Encode(
			protonostale.NewStatus(eventing.Code_UNEXPECTED_ERROR),
		)
	}
}
