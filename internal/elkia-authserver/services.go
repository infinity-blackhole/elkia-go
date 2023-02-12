package elkiaauthserver

import (
	"github.com/infinity-blackhole/elkia/internal/app"
	"github.com/infinity-blackhole/elkia/pkg/nostale"

	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

func NewNosTaleServer(c *app.Config) (*nostale.Server, error) {
	ip := app.NewIdentityProviderClient(c)
	sm, err := app.NewSessionStoreClient(c)
	if err != nil {
		return nil, err
	}
	fm, err := app.NewFleetClient(c)
	if err != nil {
		return nil, err
	}
	handler := NewHandler(HandlerConfig{
		IAMClient:    ip,
		SessionStore: sm,
		Fleet:        fm,
	})
	return nostale.NewServer(nostale.ServerConfig{
		Addr:    ":8080",
		Handler: handler,
	}), nil
}
