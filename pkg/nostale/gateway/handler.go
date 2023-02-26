package gateway

import (
	"bufio"
	"context"
	"errors"
	"net"
	"strconv"

	eventing "github.com/infinity-blackhole/elkia/pkg/api/eventing/v1alpha1"
	"github.com/infinity-blackhole/elkia/pkg/protonostale"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
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
		rc:      bufio.NewReader(NewReader(bufio.NewReader(c))),
		wc:      bufio.NewWriter(NewWriter(bufio.NewWriter(c))),
		gateway: h.gateway,
	}
}

type handoffConn struct {
	rwc     net.Conn
	rc      *bufio.Reader
	wc      *bufio.Writer
	gateway eventing.GatewayClient
}

func (c *handoffConn) serve(ctx context.Context) {
	go func() {
		if err := recover(); err != nil {
			protonostale.WriteDialogErrorEvent(c.wc, &eventing.DialogErrorEvent{
				Code: eventing.DialogErrorCode_UNEXPECTED_ERROR,
			})
		}
	}()
	ctx, span := otel.Tracer(name).Start(ctx, "Serve")
	defer span.End()
	stream, err := c.gateway.AuthHandoffInteract(ctx)
	if err != nil {
		logrus.Fatalf("gateway: handoff failed: %v", err)
	}
	conn, err := c.handleHandoff(ctx, stream)
	if err != nil {
		logrus.Debugf("gateway: handoff failed: %v", err)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return
	}
	logrus.Debugf("gateway: handoff success: %v", c.rwc.RemoteAddr())
	conn.serve(ctx)
}

func (c *handoffConn) handleHandoff(
	ctx context.Context,
	stream eventing.Gateway_AuthHandoffInteractClient,
) (*channelConn, error) {
	scanner := bufio.NewScanner(c.rc)
	scanner.Split(bufio.ScanLines)
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return nil, err
		}
		return nil, errors.New("gateway: handshake failed: no data")
	}
	sync, err := protonostale.ParseAuthHandoffSyncEvent(scanner.Text())
	if err != nil {
		if _, err := protonostale.WriteDialogErrorEvent(
			c.wc,
			&eventing.DialogErrorEvent{
				Code: eventing.DialogErrorCode_BAD_CASE,
			}); err != nil {
			return nil, err
		}
		return nil, nil
	}
	if err := stream.Send(&eventing.AuthHandoffInteractRequest{
		Payload: &eventing.AuthHandoffInteractRequest_SyncEvent{
			SyncEvent: sync,
		},
	}); err != nil {
		return nil, err
	}
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return nil, err
		}
		return nil, errors.New("gateway: handshake failed: no data")
	}
	login, err := protonostale.ParseAuthHandoffLoginEvent(scanner.Text())
	if err != nil {
		if _, err := protonostale.WriteDialogErrorEvent(
			c.wc,
			&eventing.DialogErrorEvent{
				Code: eventing.DialogErrorCode_BAD_CASE,
			}); err != nil {
			return nil, err
		}
		return nil, nil
	}
	if err := stream.Send(&eventing.AuthHandoffInteractRequest{
		Payload: &eventing.AuthHandoffInteractRequest_LoginEvent{
			LoginEvent: login,
		},
	}); err != nil {
		return nil, err
	}
	m, err := stream.Recv()
	if err != nil {
		return nil, err
	}
	ack := m.GetLoginSuccessEvent()
	if ack == nil {
		protonostale.WriteDialogErrorEvent(c.wc, &eventing.DialogErrorEvent{
			Code: eventing.DialogErrorCode_CANT_AUTHENTICATE,
		})
		return nil, err
	}
	return c.newChannelConn(ack.Key, login.PasswordEvent.Sequence), nil
}

func (h *handoffConn) newChannelConn(sequence, key uint32) *channelConn {
	return &channelConn{
		rwc: h.rwc,
		rc: bufio.NewReader(
			NewReaderWithKey(bufio.NewReader(h.rwc), key),
		),
		wc: bufio.NewWriter(
			NewWriter(bufio.NewWriter(h.rwc)),
		),
		gateway:  h.gateway,
		key:      key,
		sequence: sequence,
	}
}

type channelConn struct {
	rwc      net.Conn
	rc       *bufio.Reader
	wc       *bufio.Writer
	gateway  eventing.GatewayClient
	key      uint32
	sequence uint32
}

func (c *channelConn) serve(ctx context.Context) {
	go func() {
		if err := recover(); err != nil {
			protonostale.WriteDialogErrorEvent(c.wc, &eventing.DialogErrorEvent{
				Code: eventing.DialogErrorCode_UNEXPECTED_ERROR,
			})
		}
	}()
	_, span := otel.Tracer(name).Start(ctx, "Serve")
	defer span.End()
	ctx = metadata.AppendToOutgoingContext(
		ctx,
		"sequence", strconv.Itoa(int(c.sequence)),
		"key", strconv.Itoa(int(c.key)),
	)
	stream, err := c.gateway.ChannelInteract(ctx)
	if err != nil {
		logrus.Debugf("gateway: handoff failed: %v", err)
	}
	for {
		m, err := protonostale.ReadChannelEvent(c.rc)
		if err != nil {
			logrus.Debugf("gateway: read failed: %v", err)
			return
		}
		if err := stream.Send(&eventing.ChannelInteractRequest{
			Payload: &eventing.ChannelInteractRequest_ChannelEvent{
				ChannelEvent: m,
			},
		}); err != nil {
			logrus.Fatalf("gateway: send failed: %v", err)
		}
	}
}
