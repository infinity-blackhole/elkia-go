package elkiaauthserver

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"net/textproto"

	"github.com/infinity-blackhole/elkia/pkg/core"
	"github.com/infinity-blackhole/elkia/pkg/crypto"
	"github.com/infinity-blackhole/elkia/pkg/messages"
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
	r := NewReader(bufio.NewReader(c))
	m, err := ReadCredentialsMessage(r)
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
	key := HandoffSessionKeyFromOrySession(session)
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
	gateways := make([]messages.Gateway, len(worlds))
	for _, w := range worlds {
		gateways = append(gateways, GatewaysFromFleetWorld(&w)...)
	}
	evs := messages.ListGatewaysMessage{
		Key:      key,
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
