package authserver

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"net/textproto"

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

	listWorlds, err := h.fleet.
		ListWorlds(ctx, &fleetv1alpha1pb.ListWorldRequest{})
	if err != nil {
		panic(err)
	}
	gateways := make([]*eventingv1alpha1pb.Gateway, len(listWorlds.Worlds))
	for _, w := range listWorlds.Worlds {
		listGateways, err := h.fleet.
			ListGateways(ctx, &fleetv1alpha1pb.ListGatewayRequest{
				WorldId: w.Id,
			})
		if err != nil {
			panic(err)
		}
		for _, g := range listGateways.Gateways {
			gateways = append(
				gateways,
				&eventingv1alpha1pb.Gateway{
					Address:    g.Address,
					Population: g.Population,
					Capacity:   g.Capacity,
					WorldId:    w.Id,
					Id:         g.Id,
					WorldName:  w.Name,
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
	r      *textproto.Reader
	crypto *crypto.SimpleSubstitution
}

func NewReader(r *bufio.Reader) *Reader {
	return &Reader{
		r:      textproto.NewReader(r),
		crypto: new(crypto.SimpleSubstitution),
	}
}

func (r *Reader) ReadLine() ([]byte, error) {
	s, err := r.r.ReadLineBytes()
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
