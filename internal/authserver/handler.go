package authserver

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

const name = "github.com/infinity-blackhole/elkia/internal/authserver"

type HandlerConfig struct {
	FleetClient fleet.FleetClient
}

func NewHandler(cfg HandlerConfig) *Handler {
	return &Handler{
		fleetClient: cfg.FleetClient,
	}
}

type Handler struct {
	fleetClient fleet.FleetClient
}

func (h *Handler) ServeNosTale(c net.Conn) {
	ctx := context.Background()
	conn := h.newConn(c)
	logrus.Debugf("authserver: new connection from %s", c.RemoteAddr().String())
	go conn.serve(ctx)
}

func (h *Handler) newConn(c net.Conn) *Conn {
	return &Conn{
		rwc:         c,
		rc:          protonostale.NewAuthServerReader(bufio.NewReader(c)),
		wc:          protonostale.NewAuthServerWriter(bufio.NewWriter(c)),
		fleetClient: h.fleetClient,
	}
}

type Conn struct {
	rwc         net.Conn
	rc          *protonostale.AuthServerReader
	wc          *protonostale.AuthServerWriter
	fleetClient fleet.FleetClient
}

func (c *Conn) serve(ctx context.Context) {
	ctx, span := otel.Tracer(name).Start(ctx, "Serve")
	defer span.End()
	r, err := c.rc.ReadMessage()
	if err != nil {
		if err := c.wc.WriteFailCodeMessage(&eventing.FailureMessage{
			Code: eventing.FailureCode_UNEXPECTED_ERROR,
		}); err != nil {
			logrus.Fatal(err)
		}
		logrus.Debug(err)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return
	}
	opcode, err := r.ReadOpcode()
	if err != nil {
		if err := c.wc.WriteFailCodeMessage(&eventing.FailureMessage{
			Code: eventing.FailureCode_BAD_CASE,
		}); err != nil {
			logrus.Fatal(err)
		}
		logrus.Debug(err)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return
	}
	logrus.Debugf("authserver: read opcode: %v", opcode)
	switch opcode {
	case protonostale.HandoffOpCode:
		logrus.Debugf("authserver: handle handoff")
		c.handleHandoff(ctx, r)
	default:
		c.handleFallback(opcode)
	}
}

func (c *Conn) handleHandoff(ctx context.Context, r *protonostale.AuthServerMessageReader) {
	ctx, span := otel.Tracer(name).Start(ctx, "Handle Handoff")
	defer span.End()
	m, err := r.ReadRequestHandoffMessage()
	if err != nil {
		if err := c.wc.WriteFailCodeMessage(&eventing.FailureMessage{
			Code: eventing.FailureCode_BAD_CASE,
		}); err != nil {
			logrus.Fatal(err)
		}
		logrus.Debug(err)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return
	}
	logrus.Debugf("authserver: read handoff: %v", m)

	handoff, err := c.fleetClient.
		CreateHandoff(
			ctx,
			&fleet.CreateHandoffRequest{
				Identifier: m.Identifier,
				Token:      m.Password,
			},
		)
	if err != nil {
		if err := c.wc.WriteFailCodeMessage(&eventing.FailureMessage{
			Code: eventing.FailureCode_UNEXPECTED_ERROR,
		}); err != nil {
			logrus.Fatal(err)
		}
		logrus.Debug(err)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return
	}
	logrus.Debugf("authserver: create handoff: %v", handoff)

	listClusters, err := c.fleetClient.
		ListClusters(ctx, &fleet.ListClusterRequest{})
	if err != nil {
		if err := c.wc.WriteFailCodeMessage(&eventing.FailureMessage{
			Code: eventing.FailureCode_UNEXPECTED_ERROR,
		}); err != nil {
			logrus.Fatal(err)
		}
		logrus.Debug(err)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return
	}
	logrus.Debugf("authserver: list clusters: %v", listClusters)
	gateways := []*eventing.GatewayMessage{}
	for _, cluster := range listClusters.Clusters {
		listGateways, err := c.fleetClient.
			ListGateways(ctx, &fleet.ListGatewayRequest{
				Id: cluster.Id,
			})
		if err != nil {
			if err := c.wc.WriteFailCodeMessage(&eventing.FailureMessage{
				Code: eventing.FailureCode_UNEXPECTED_ERROR,
			}); err != nil {
				logrus.Fatal(err)
			}
			logrus.Debug(err)
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return
		}
		logrus.Debugf("authserver: list gateways: %v", listGateways)
		for _, g := range listGateways.Gateways {
			host, port, err := net.SplitHostPort(g.Address)
			if err != nil {
				if err := c.wc.WriteFailCodeMessage(&eventing.FailureMessage{
					Code: eventing.FailureCode_UNEXPECTED_ERROR,
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
				&eventing.GatewayMessage{
					Host:       host,
					Port:       port,
					Population: g.Population,
					Capacity:   g.Capacity,
					WorldId:    cluster.WorldId,
					ChannelId:  g.ChannelId,
					WorldName:  cluster.Name,
				},
			)
		}
	}
	if err := c.wc.WriteProposeHandoffMessage(
		&eventing.ProposeHandoffMessage{
			Key:      handoff.Key,
			Gateways: gateways,
		},
	); err != nil {
		if err := c.wc.WriteFailCodeMessage(&eventing.FailureMessage{
			Code: eventing.FailureCode_UNEXPECTED_ERROR,
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
	if err := c.wc.WriteFailCodeMessage(&eventing.FailureMessage{
		Code: eventing.FailureCode_BAD_CASE,
	}); err != nil {
		logrus.Fatal(err)
	}
	logrus.Debugf("unknown opcode: %s", opcode)
}
