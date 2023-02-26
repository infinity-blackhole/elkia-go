package gateway

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"strconv"

	eventing "github.com/infinity-blackhole/elkia/pkg/api/eventing/v1alpha1"
	"github.com/infinity-blackhole/elkia/pkg/nostale/utils"
	"github.com/infinity-blackhole/elkia/pkg/protonostale"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
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
			utils.WriteError(
				c.wc,
				eventing.DialogErrorCode_UNEXPECTED_ERROR,
				"An unexpected error occurred, please try again later",
			)
		}
	}()
	c.handleAuthHandoff(ctx)
}

func (c *handoffConn) handleAuthHandoff(ctx context.Context) {
	_, span := otel.Tracer(name).Start(ctx, "Handle Handoff")
	defer span.End()
	stream, err := c.gateway.AuthHandoffInteract(ctx)
	if err != nil {
		utils.WriteError(
			c.wc,
			eventing.DialogErrorCode_UNEXPECTED_ERROR,
			fmt.Sprintf(
				"failed to create auth handoff interact stream: %v",
				err,
			),
		)
		return
	}
	scanner := bufio.NewScanner(c.rc)
	scanner.Split(bufio.ScanLines)
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			utils.WriteError(
				c.wc,
				eventing.DialogErrorCode_UNEXPECTED_ERROR,
				fmt.Sprintf(
					"failed to read handshake event: %v",
					err,
				),
			)
			return
		}
		utils.WriteError(
			c.wc,
			eventing.DialogErrorCode_UNEXPECTED_ERROR,
			"failed to read handshake event: EOF",
		)
		return
	}
	sync, err := protonostale.ParseAuthHandoffSyncEvent(scanner.Text())
	if err != nil {
		utils.WriteError(
			c.wc,
			eventing.DialogErrorCode_BAD_CASE,
			fmt.Sprintf(
				"failed to parse auth handoff sync event: %v",
				err,
			),
		)
		return
	}
	if err := stream.Send(&eventing.AuthHandoffInteractRequest{
		Payload: &eventing.AuthHandoffInteractRequest_SyncEvent{
			SyncEvent: sync,
		},
	}); err != nil {
		utils.WriteError(
			c.wc,
			eventing.DialogErrorCode_UNEXPECTED_ERROR,
			fmt.Sprintf(
				"failed to send auth handoff sync event: %v",
				err,
			),
		)
		return
	}
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			utils.WriteError(
				c.wc,
				eventing.DialogErrorCode_BAD_CASE,
				fmt.Sprintf(
					"failed to read auth handoff login event: %v",
					err,
				),
			)
			return
		}
		utils.WriteError(
			c.wc,
			eventing.DialogErrorCode_BAD_CASE,
			"failed to read auth handoff login event: EOF",
		)
		return
	}
	login, err := protonostale.ParseAuthHandoffLoginEvent(scanner.Text())
	if err != nil {
		utils.WriteError(
			c.wc,
			eventing.DialogErrorCode_BAD_CASE,
			fmt.Sprintf(
				"failed to parse auth handoff login event: %v",
				err,
			),
		)
		return
	}
	if err := stream.Send(&eventing.AuthHandoffInteractRequest{
		Payload: &eventing.AuthHandoffInteractRequest_LoginEvent{
			LoginEvent: login,
		},
	}); err != nil {
		utils.WriteError(
			c.wc,
			eventing.DialogErrorCode_UNEXPECTED_ERROR,
			fmt.Sprintf(
				"failed to send auth handoff login event: %v",
				err,
			),
		)
		return
	}
	m, err := stream.Recv()
	if err != nil {
		utils.WriteError(
			c.wc,
			eventing.DialogErrorCode_UNEXPECTED_ERROR,
			fmt.Sprintf(
				"failed to receive auth handoff login success event: %v",
				err,
			),
		)
		return
	}
	ack := m.GetLoginSuccessEvent()
	if ack == nil {
		utils.WriteError(
			c.wc,
			eventing.DialogErrorCode_BAD_CASE,
			fmt.Sprintf(
				"failed to receive auth handoff login success event: %v",
				err,
			),
		)
		return
	}
	conn := c.newChannelConn(ack.Key, login.PasswordEvent.Sequence)
	go conn.serve(ctx)
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
			utils.WriteError(
				c.wc,
				eventing.DialogErrorCode_UNEXPECTED_ERROR,
				"An unexpected error occurred, please try again later",
			)
		}
	}()
	c.handleMessages(ctx)
}

func (c *channelConn) handleMessages(ctx context.Context) {
	_, span := otel.Tracer(name).Start(ctx, "Handle Messages")
	defer span.End()
	ctx = metadata.AppendToOutgoingContext(
		ctx,
		"sequence", strconv.FormatUint(uint64(c.sequence), 10),
		"key", strconv.FormatUint(uint64(c.key), 10),
	)
	stream, err := c.gateway.ChannelInteract(ctx)
	if err != nil {
		utils.WriteError(
			c.wc,
			eventing.DialogErrorCode_UNEXPECTED_ERROR,
			fmt.Sprintf(
				"failed to create channel interact stream: %v",
				err,
			),
		)
	}
	for {
		m, err := protonostale.ReadChannelEvent(c.rc)
		if err != nil {
			utils.WriteError(
				c.wc,
				eventing.DialogErrorCode_UNEXPECTED_ERROR,
				fmt.Sprintf(
					"failed to read channel event: %v",
					err,
				),
			)
		}
		if err := stream.Send(&eventing.ChannelInteractRequest{
			Payload: &eventing.ChannelInteractRequest_ChannelEvent{
				ChannelEvent: m,
			},
		}); err != nil {
			utils.WriteError(
				c.wc,
				eventing.DialogErrorCode_UNEXPECTED_ERROR,
				fmt.Sprintf(
					"failed to send channel event: %v",
					err,
				),
			)
		}
	}
}
