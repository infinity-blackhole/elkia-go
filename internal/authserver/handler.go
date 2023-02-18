package authserver

import (
	"bufio"
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
	rc := simplesubtitution.NewReader(bufio.NewReader(c))
	wc := simplesubtitution.NewWriter(bufio.NewWriter(c))
	m, err := ReadCredentialsMessage(rc)
	if err != nil {
		panic(err)
	}

	handoff, err := h.fleet.
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

	listClusters, err := h.fleet.
		ListClusters(ctx, &fleet.ListClusterRequest{})
	if err != nil {
		panic(err)
	}
	gateways := []*eventing.Gateway{}
	for _, c := range listClusters.Clusters {
		listGateways, err := h.fleet.
			ListGateways(ctx, &fleet.ListGatewayRequest{
				Id: c.Id,
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
				&eventing.Gateway{
					Host:       host,
					Port:       port,
					Population: g.Population,
					Capacity:   g.Capacity,
					WorldId:    c.WorldId,
					ChannelId:  g.ChannelId,
					WorldName:  c.Name,
				},
			)
		}
	}
	if err := wc.WriteMessage(protonostale.MarshallProposeHandoffMessage(
		&eventing.ProposeHandoffMessage{
			Key:      handoff.Key,
			Gateways: gateways,
		},
	)); err != nil {
		panic(err)
	}
}

func ReadCredentialsMessage(
	r *simplesubtitution.Reader,
) (*eventing.RequestHandoffMessage, error) {
	s, err := r.ReadMessageBytes()
	if err != nil {
		return nil, err
	}
	return protonostale.ParseRequestHandoffMessage(s)
}
