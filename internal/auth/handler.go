package auth

import (
	"bufio"
	"context"
	"net"

	eventing "github.com/infinity-blackhole/elkia/pkg/api/eventing/v1alpha1"
	fleet "github.com/infinity-blackhole/elkia/pkg/api/fleet/v1alpha1"
	"github.com/infinity-blackhole/elkia/pkg/protonostale"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
)

const name = "github.com/infinity-blackhole/elkia/internal/auth"

type HandlerConfig struct {
	PresenceClient fleet.PresenceClient
	ClusterClient  fleet.ClusterClient
}

func NewHandler(cfg HandlerConfig) *Handler {
	return &Handler{
		presence: cfg.PresenceClient,
		cluster:  cfg.ClusterClient,
	}
}

type Handler struct {
	presence fleet.PresenceClient
	cluster  fleet.ClusterClient
}

func (h *Handler) ServeNosTale(c net.Conn) {
	ctx := context.Background()
	conn := h.newConn(c)
	logrus.Debugf("auth: new connection from %s", c.RemoteAddr().String())
	go conn.serve(ctx)
}

func (h *Handler) newConn(c net.Conn) *Conn {
	return &Conn{
		rwc:      c,
		rc:       protonostale.NewAuthReader(bufio.NewReader(c)),
		wc:       protonostale.NewAuthWriter(bufio.NewWriter(c)),
		presence: h.presence,
		cluster:  h.cluster,
	}
}

type Conn struct {
	rwc      net.Conn
	rc       *protonostale.AuthReader
	wc       *protonostale.AuthWriter
	presence fleet.PresenceClient
	cluster  fleet.ClusterClient
}

func (c *Conn) serve(ctx context.Context) {
	ctx, span := otel.Tracer(name).Start(ctx, "Serve")
	defer span.End()
	r, err := c.rc.ReadMessage()
	if err != nil {
		if err := c.wc.WriteErrorMessageEvent(&eventing.ErrorMessageEvent{
			Code: eventing.ErrorMessageCode_UNEXPECTED_ERROR,
		}); err != nil {
			logrus.Fatal(err)
		}
		logrus.Debug(err)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return
	}
	opcode, err := r.ReadOpCode()
	if err != nil {
		if err := c.wc.WriteErrorMessageEvent(&eventing.ErrorMessageEvent{
			Code: eventing.ErrorMessageCode_BAD_CASE,
		}); err != nil {
			logrus.Fatal(err)
		}
		logrus.Debug(err)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return
	}
	logrus.Debugf("auth: read opcode: %v", opcode)
	switch opcode {
	case protonostale.HandoffOpCode:
		logrus.Debugf("auth: handle handoff")
		c.handleHandoff(ctx, r)
	default:
		logrus.Debugf("auth: handle fallback")
		c.handleFallback(opcode)
	}
}

func (c *Conn) handleHandoff(ctx context.Context, r *protonostale.AuthEventReader) {
	ctx, span := otel.Tracer(name).Start(ctx, "Handle Handoff")
	defer span.End()
	m, err := r.ReadAuthLoginEvent()
	if err != nil {
		if err := c.wc.WriteErrorMessageEvent(&eventing.ErrorMessageEvent{
			Code: eventing.ErrorMessageCode_BAD_CASE,
		}); err != nil {
			logrus.Fatal(err)
		}
		logrus.Debug(err)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return
	}
	logrus.Debugf("auth: read handoff: %v", m)

	handoff, err := c.presence.AuthLogin(
		ctx,
		&fleet.AuthLoginRequest{
			Identifier: m.Identifier,
			Password:   m.Password,
		},
	)
	if err != nil {
		if err := c.wc.WriteErrorMessageEvent(&eventing.ErrorMessageEvent{
			Code: eventing.ErrorMessageCode_UNEXPECTED_ERROR,
		}); err != nil {
			logrus.Fatal(err)
		}
		logrus.Debug(err)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return
	}
	logrus.Debugf("auth: create handoff: %v", handoff)

	MemberList, err := c.cluster.MemberList(ctx, &fleet.MemberListRequest{})
	if err != nil {
		if err := c.wc.WriteErrorMessageEvent(&eventing.ErrorMessageEvent{
			Code: eventing.ErrorMessageCode_UNEXPECTED_ERROR,
		}); err != nil {
			logrus.Fatal(err)
		}
		logrus.Debug(err)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return
	}
	logrus.Debugf("auth: list members: %v", MemberList)
	gateways := []*eventing.Gateway{}
	for _, m := range MemberList.Members {
		host, port, err := net.SplitHostPort(m.Address)
		if err != nil {
			if err := c.wc.WriteErrorMessageEvent(&eventing.ErrorMessageEvent{
				Code: eventing.ErrorMessageCode_UNEXPECTED_ERROR,
			}); err != nil {
				logrus.Fatal(err)
			}
			logrus.Debug(err)
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return
		}
		gateways = append(
			gateways,
			&eventing.Gateway{
				Host:       host,
				Port:       port,
				Population: m.Population,
				Capacity:   m.Capacity,
				WorldId:    m.WorldId,
				ChannelId:  m.ChannelId,
				WorldName:  m.Name,
			},
		)
	}
	if err := c.wc.WriteGatewayListEvent(
		&eventing.GatewayListEvent{
			Key:      handoff.Key,
			Gateways: gateways,
		},
	); err != nil {
		if err := c.wc.WriteErrorMessageEvent(&eventing.ErrorMessageEvent{
			Code: eventing.ErrorMessageCode_UNEXPECTED_ERROR,
		}); err != nil {
			logrus.Fatal(err)
		}
		logrus.Debug(err)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return
	}
}

func (c *Conn) handleFallback(opcode string) {
	if err := c.wc.WriteErrorMessageEvent(&eventing.ErrorMessageEvent{
		Code: eventing.ErrorMessageCode_BAD_CASE,
	}); err != nil {
		logrus.Fatal(err)
	}
	logrus.Debugf("unknown opcode: %s", opcode)
}
