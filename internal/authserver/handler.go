package authserver

import (
	"bufio"
	"context"
	"fmt"
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
	r := NewReader(bufio.NewReader(c))
	m, err := ReadCredentialsMessage(r)
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
	gateways := make([]*eventingv1alpha1pb.Gateway, len(listClusters.Clusters))
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
	fmt.Fprint(c, eventingv1alpha1pb.ProposeHandoffMessage{
		Key:      handoff.Key,
		Gateways: gateways,
	})
}

type Reader struct {
	r      *bufio.Reader
	crypto *crypto.SimpleSubstitution
}

func NewReader(r *bufio.Reader) *Reader {
	return &Reader{
		r:      r,
		crypto: new(crypto.SimpleSubstitution),
	}
}

func (r *Reader) ReadLine() ([]byte, error) {
	s, err := r.r.ReadBytes(0xD8)
	if err != nil {
		return nil, err
	}
	return r.crypto.Decrypt(s), nil
}

func ReadCredentialsMessage(
	r *Reader,
) (*eventingv1alpha1pb.RequestHandoffMessage, error) {
	s, err := r.ReadLine()
	if err != nil {
		return nil, err
	}
	return protonostale.ParseRequestHandoffMessage(s)
}
