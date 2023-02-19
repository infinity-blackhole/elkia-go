package authserver

import (
	"bufio"
	"context"
	"net"

	eventing "github.com/infinity-blackhole/elkia/pkg/api/eventing/v1alpha1"
	fleet "github.com/infinity-blackhole/elkia/pkg/api/fleet/v1alpha1"
	"github.com/infinity-blackhole/elkia/pkg/protonostale"
	"github.com/sirupsen/logrus"
)

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
	logrus.Debugf("start serving %v", c.rwc.RemoteAddr())
	r, err := c.rc.ReadMessage()
	if err != nil {
		if err := c.wc.WriteFailCodeMessage(&eventing.FailureMessage{
			Code: eventing.FailureCode_UNEXPECTED_ERROR,
		}); err != nil {
			logrus.Fatal(err)
		}
		logrus.Debug(err)
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
		return
	}
	logrus.Debugf("read opcode: %s", opcode)
	switch opcode {
	case protonostale.HandoffOpCode:
		c.handleHandoff(ctx, r)
	default:
		c.handleFallback(opcode)
	}
}

func (c *Conn) handleHandoff(ctx context.Context, r *protonostale.AuthServerMessageReader) {
	logrus.Debug("handle handoff")
	m, err := r.ReadRequestHandoffMessage()
	if err != nil {
		if err := c.wc.WriteFailCodeMessage(&eventing.FailureMessage{
			Code: eventing.FailureCode_BAD_CASE,
		}); err != nil {
			logrus.Fatal(err)
		}
		logrus.Debug(err)
		return
	}
	logrus.Debugf("read request handoff message: %v", m)

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
		return
	}
	logrus.Debugf("create handoff: %v", handoff)

	listClusters, err := c.fleetClient.
		ListClusters(ctx, &fleet.ListClusterRequest{})
	if err != nil {
		if err := c.wc.WriteFailCodeMessage(&eventing.FailureMessage{
			Code: eventing.FailureCode_UNEXPECTED_ERROR,
		}); err != nil {
			logrus.Fatal(err)
		}
		logrus.Debug(err)
		return
	}
	logrus.Debugf("list clusters: %v", listClusters)
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
			return
		}
		logrus.Debugf("list gateways: %v", listGateways)
		for _, g := range listGateways.Gateways {
			host, port, err := net.SplitHostPort(g.Address)
			if err != nil {
				if err := c.wc.WriteFailCodeMessage(&eventing.FailureMessage{
					Code: eventing.FailureCode_UNEXPECTED_ERROR,
				}); err != nil {
					logrus.Fatal(err)
				}
				logrus.Debug(err)
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
