package fleet

import (
	"context"

	fleet "github.com/infinity-blackhole/elkia/pkg/api/fleet/v1alpha1"
	ory "github.com/ory/client-go"
)

type IdentityProviderServiceConfig struct {
	OryClient *ory.APIClient
}

func NewIdentityProvider(config *IdentityProviderServiceConfig) *IdentityProvider {
	return &IdentityProvider{
		oryClient: config.OryClient,
	}
}

type IdentityProvider struct {
	oryClient *ory.APIClient
}

func (i *IdentityProvider) PerformLoginFlowWithPasswordMethod(
	ctx context.Context,
	identifier, password string,
) (*fleet.Session, error) {
	flow, _, err := i.oryClient.FrontendApi.
		CreateNativeLoginFlow(ctx).
		Execute()
	if err != nil {
		return nil, err
	}
	ulf, _, err := i.oryClient.FrontendApi.
		UpdateLoginFlow(ctx).
		Flow(flow.Id).
		UpdateLoginFlowBody(
			ory.UpdateLoginFlowWithPasswordMethodAsUpdateLoginFlowBody(
				ory.NewUpdateLoginFlowWithPasswordMethod(
					identifier,
					"password",
					password,
				),
			),
		).
		Execute()
	if err != nil {
		return nil, err
	}
	return &fleet.Session{
		Id:         ulf.Session.Id,
		IdentityId: ulf.Session.Identity.Id,
		Token:      *ulf.SessionToken,
	}, nil
}

func (i *IdentityProvider) GetSession(
	ctx context.Context, token string,
) (*fleet.Session, error) {
	session, _, err := i.oryClient.FrontendApi.
		ToSession(ctx).
		XSessionToken(token).
		Execute()
	if err != nil {
		return nil, err
	}
	return &fleet.Session{
		Id:         session.Id,
		IdentityId: session.Identity.Id,
		Token:      token,
	}, nil
}

func (i *IdentityProvider) Logout(ctx context.Context, sessionToken string) error {
	_, err := i.oryClient.FrontendApi.PerformNativeLogout(ctx).
		PerformNativeLogoutBody(
			*ory.NewPerformNativeLogoutBody(sessionToken),
		).
		Execute()
	return err
}
