package authserver

import (
	"bufio"
	"context"
	"fmt"
	"hash/fnv"
	"net"
	"net/textproto"

	fleetv1alpha1pb "github.com/infinity-blackhole/elkia/pkg/api/fleet/v1alpha1"
	"github.com/infinity-blackhole/elkia/pkg/crypto"
	"github.com/infinity-blackhole/elkia/pkg/messages"
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
	gateways := make([]*messages.Gateway, len(listWorlds.Worlds))
	for _, w := range listWorlds.Worlds {
		listGateways, err := h.fleet.
			ListGateways(ctx, &fleetv1alpha1pb.ListGatewayRequest{
				WorldId: w.Id,
			})
		if err != nil {
			panic(err)
		}
		for _, g := range listGateways.Gateways {
			gateways = append(gateways, GatewayFromFleetGateway(w.Id, w.Name, g))
		}
	}
	evs := messages.ListGatewaysMessage{
		Key:      handoff.Key,
		Gateways: gateways,
	}
	fmt.Fprint(c, evs)
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

func ReadCredentialsMessage(r *Reader) (*messages.CredentialsMessage, error) {
	s, err := r.ReadLine()
	if err != nil {
		return nil, err
	}
	return messages.ParseCredentialsMessage(s)
}

func GatewayFromFleetGateway(id, name string, g *fleetv1alpha1pb.Gateway) *messages.Gateway {
	h := fnv.New32a()
	h.Write([]byte(id))
	worldIdNum := h.Sum32()
	h = fnv.New32a()
	h.Write([]byte(g.Id))
	gatewayId := h.Sum32()
	return &messages.Gateway{
		Addr:       g.Address,
		Population: g.Population,
		Capacity:   g.Capacity,
		WorldID:    worldIdNum,
		ID:         gatewayId,
		WorldName:  name,
	}
}
