package elkiaauthserver

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"net/textproto"

	"github.com/infinity-blackhole/elkia/pkg/core"
)

type HandlerConfig struct {
	IAMClient    *core.IdentityProvider
	SessionStore *core.SessionStore
	Fleet        *core.FleetClient
}

func NewHandler(cfg HandlerConfig) *Handler {
	return &Handler{
		identityProvider: cfg.IAMClient,
		sessionStore:     cfg.SessionStore,
		fleet:            cfg.Fleet,
	}
}

type Handler struct {
	identityProvider *core.IdentityProvider
	sessionStore     *core.SessionStore
	fleet            *core.FleetClient
}

func (h *Handler) ServeNosTale(c net.Conn) {
	ctx := context.Background()
	r := textproto.NewReader(bufio.NewReader(c))
	s, err := r.ReadLine()
	if err != nil {
		panic(err)
	}
	m, err := ParseCredentialsMessage(s)
	if err != nil {
		panic(err)
	}

	session, token, err := h.identityProvider.PerformLoginFlowWithPasswordMethod(
		ctx,
		m.Identifier,
		m.Password,
	)
	if err != nil {
		panic(err)
	}
	key := SessionKeyFromOrySession(session)
	if err := h.sessionStore.SetHandoffSession(ctx, key, &core.HandoffSession{
		Identifier:   m.Identifier,
		SessionToken: token,
	}); err != nil {
		panic(err)
	}

	worlds, err := h.fleet.ListWorlds()
	if err != nil {
		panic(err)
	}
	gateways := make([]Gateway, len(worlds))
	for _, w := range worlds {
		gateways = append(gateways, GatewaysFromFleetWorld(&w)...)
	}
	evs := ListGatewaysMessage{
		Key:      key,
		Gateways: gateways,
	}
	fmt.Fprint(c, evs)
}
