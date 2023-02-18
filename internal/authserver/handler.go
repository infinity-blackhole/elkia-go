package authserver

import (
	"bufio"
	"context"
	"net"

	eventing "github.com/infinity-blackhole/elkia/pkg/api/eventing/v1alpha1"
	fleet "github.com/infinity-blackhole/elkia/pkg/api/fleet/v1alpha1"
	"github.com/infinity-blackhole/elkia/pkg/protonostale"
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
	r, err := c.rc.ReadMessage()
	if err != nil {
		if err := c.wc.WriteFailCodeMessage(&eventing.FailureMessage{
			Code: eventing.FailureCode_UNEXPECTED_ERROR,
		}); err != nil {
			panic(err)
		}
		return
	}
	m, err := r.ReadRequestHandoffMessage()
	if err != nil {
		if err := c.wc.WriteFailCodeMessage(&eventing.FailureMessage{
			Code: eventing.FailureCode_BAD_CASE,
		}); err != nil {
			panic(err)
		}
		return
	}
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
			panic(err)
		}
		return
	}

	listClusters, err := c.fleetClient.
		ListClusters(ctx, &fleet.ListClusterRequest{})
	if err != nil {
		if err := c.wc.WriteFailCodeMessage(&eventing.FailureMessage{
			Code: eventing.FailureCode_UNEXPECTED_ERROR,
		}); err != nil {
			panic(err)
		}
		return
	}
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
				panic(err)
			}
			return
		}
		for _, g := range listGateways.Gateways {
			host, port, err := net.SplitHostPort(g.Address)
			if err != nil {
				if err := c.wc.WriteFailCodeMessage(&eventing.FailureMessage{
					Code: eventing.FailureCode_UNEXPECTED_ERROR,
				}); err != nil {
					panic(err)
				}
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
			panic(err)
		}
		return
	}
}
