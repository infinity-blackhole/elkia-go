package elkiaauthserver

import (
	"encoding/gob"
	"hash/fnv"

	"github.com/infinity-blackhole/elkia/pkg/core"
	"github.com/infinity-blackhole/elkia/pkg/messages"
	ory "github.com/ory/client-go"
)

func HandoffSessionKeyFromOrySession(session *ory.Session) uint32 {
	h := fnv.New32a()
	if err := gob.
		NewEncoder(h).
		Encode(session.Id); err != nil {
		panic(err)
	}
	return h.Sum32()
}

func GatewaysFromFleetWorld(world *core.World) []messages.Gateway {
	gateways := make([]messages.Gateway, len(world.Gateways))
	for _, g := range world.Gateways {
		gateways = append(gateways, GatewayFromFleetGateway(world.ID, world.Name, &g))
	}
	return gateways
}

func GatewayFromFleetGateway(id, name string, g *core.Gateway) messages.Gateway {
	h := fnv.New32a()
	h.Write([]byte(id))
	worldIdNum := h.Sum32()
	h = fnv.New32a()
	h.Write([]byte(g.ID))
	gatewayId := h.Sum32()
	return messages.Gateway{
		Addr:       g.Addr,
		Population: g.Population,
		Capacity:   g.Capacity,
		WorldID:    worldIdNum,
		ID:         gatewayId,
		WorldName:  name,
	}
}
