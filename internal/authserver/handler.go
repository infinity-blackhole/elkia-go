package authserver

import (
	"bufio"
	"context"
	"net"

	eventingv1alpha1pb "github.com/infinity-blackhole/elkia/pkg/api/eventing/v1alpha1"
	fleetv1alpha1pb "github.com/infinity-blackhole/elkia/pkg/api/fleet/v1alpha1"
	"github.com/infinity-blackhole/elkia/pkg/crypto"
	"github.com/infinity-blackhole/elkia/pkg/protonostale"
)

type HandlerConfig struct {
	FleetClient fleetv1alpha1pb.FleetServiceClient
}

func NewHandler(cfg HandlerConfig) *Handler {
	return &Handler{
		fleet: cfg.FleetClient,
	}
}

type Handler struct {
	fleet fleetv1alpha1pb.FleetServiceClient
}

func (h *Handler) ServeNosTale(c net.Conn) {
	ctx := context.Background()
	rc := crypto.NewServerReader(bufio.NewReader(c))
	wc := crypto.NewServerWriter(bufio.NewWriter(c))
	m, err := ReadCredentialsMessage(rc)
	if err != nil {
		panic(err)
	}

	handoff, err := h.fleet.
		CreateHandoff(
			ctx,
			&fleetv1alpha1pb.CreateHandoffRequest{
				Identifier: m.Identifier,
				Token:      m.Password,
			},
		)
	if err != nil {
		panic(err)
	}

	listClusters, err := h.fleet.
		ListClusters(ctx, &fleetv1alpha1pb.ListClusterRequest{})
	if err != nil {
		panic(err)
	}
	gateways := []*eventingv1alpha1pb.Gateway{}
	for _, c := range listClusters.Clusters {
		listGateways, err := h.fleet.
			ListGateways(ctx, &fleetv1alpha1pb.ListGatewayRequest{
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
				&eventingv1alpha1pb.Gateway{
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
	if err := wc.WriteLine(protonostale.MarshallProposeHandoffMessage(
		&eventingv1alpha1pb.ProposeHandoffMessage{
			Key:      handoff.Key,
			Gateways: gateways,
		},
	)); err != nil {
		panic(err)
	}
}

func ReadCredentialsMessage(
	r *crypto.ServerReader,
) (*eventingv1alpha1pb.RequestHandoffMessage, error) {
	s, err := r.ReadLineBytes()
	if err != nil {
		return nil, err
	}
	return protonostale.ParseRequestHandoffMessage(s)
}
