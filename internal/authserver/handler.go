package authserver

import (
	"bufio"
	"bytes"
	"context"
	"net"

	eventing "github.com/infinity-blackhole/elkia/pkg/api/eventing/v1alpha1"
	fleet "github.com/infinity-blackhole/elkia/pkg/api/fleet/v1alpha1"
	"github.com/infinity-blackhole/elkia/pkg/nostale/simplesubtitution"
	"github.com/infinity-blackhole/elkia/pkg/protonostale"
)

type HandlerConfig struct {
	FleetClient fleet.FleetClient
}

func NewHandler(cfg HandlerConfig) *Handler {
	return &Handler{
		fleet: cfg.FleetClient,
	}
}

type Handler struct {
	fleet fleet.FleetClient
}

func (h *Handler) ServeNosTale(c net.Conn) {
	ctx := context.Background()
	conn := h.newConn(c)
	go conn.serve(ctx)
}

func (h *Handler) newConn(c net.Conn) *Conn {
	return &Conn{
		conn: c,
		rc:   simplesubtitution.NewReader(bufio.NewReader(c)),
	}
}

type Conn struct {
	conn        net.Conn
	wc          *simplesubtitution.Writer
	rc          *simplesubtitution.Reader
	fleetClient fleet.FleetClient
}

func (c *Conn) serve(ctx context.Context) {
	r, err := c.newMessageReader()
	if err != nil {
		panic(err)
	}
	m, err := r.ReadRequestHandoffMessage()
	if err != nil {
		panic(err)
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
		panic(err)
	}

	listClusters, err := c.fleetClient.
		ListClusters(ctx, &fleet.ListClusterRequest{})
	if err != nil {
		panic(err)
	}
	gateways := []*eventing.GatewayMessage{}
	for _, cluster := range listClusters.Clusters {
		listGateways, err := c.fleetClient.
			ListGateways(ctx, &fleet.ListGatewayRequest{
				Id: cluster.Id,
			})
		if err != nil {
			panic(err)
		}
		for _, g := range listGateways.Gateways {
			host, port, err := net.SplitHostPort(g.Address)
			if err != nil {
				panic(err)
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
	if err := c.wc.WriteMessage(protonostale.MarshallProposeHandoffMessage(
		&eventing.ProposeHandoffMessage{
			Key:      handoff.Key,
			Gateways: gateways,
		},
	)); err != nil {
		panic(err)
	}
}

func (c *Conn) newMessageReader() (*protonostale.AuthServerMessageReader, error) {
	buff, err := c.rc.ReadMessageBytes()
	if err != nil {
		return nil, err
	}
	return protonostale.NewAuthServerMessageReader(bytes.NewReader(buff)), nil
}
