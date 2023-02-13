package apiserver

import (
	"context"

	ory "github.com/ory/client-go"
)

type IdentityProviderServiceConfig struct {
	OryClient *ory.APIClient
}

func NewIdentityProviderService(config *IdentityProviderServiceConfig) *IdentityProvider {
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
) (*ory.Session, string, error) {
	flow, _, err := i.oryClient.FrontendApi.
		CreateNativeLoginFlow(ctx).
		Execute()
	if err != nil {
		return nil, "", err
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
		return nil, "", err
	}
	return &ulf.Session, *ulf.SessionToken, nil
}

func (i *IdentityProvider) GetSession(
	ctx context.Context, token string,
) (*ory.Session, error) {
	session, _, err := i.oryClient.FrontendApi.
		ToSession(ctx).
		XSessionToken(token).
		Execute()
	if err != nil {
		return nil, err
	}
	return session, nil
}

func (i *IdentityProvider) Logout(ctx context.Context, sessionToken string) error {
	_, err := i.oryClient.FrontendApi.PerformNativeLogout(ctx).
		PerformNativeLogoutBody(
			*ory.NewPerformNativeLogoutBody(sessionToken),
		).
		Execute()
	return err
}
