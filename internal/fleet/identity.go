package fleet

import (
	"context"

	fleet "github.com/infinity-blackhole/elkia/pkg/api/fleet/v1alpha1"
	ory "github.com/ory/client-go"
	"github.com/sirupsen/logrus"
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

func (i *IdentityProvider) PerformGatewayLoginFlowWithPasswordMethod(
	ctx context.Context,
	identifier, password, token string,
) (*fleet.Session, error) {
	flow, _, err := i.oryClient.FrontendApi.
		CreateNativeLoginFlow(ctx).
		XSessionToken(token).
		Refresh(true).
		Execute()
	if err != nil {
		return nil, err
	}
	logrus.Debugf("fleet: created native login flow: %v", flow)
	successLogin, _, err := i.oryClient.FrontendApi.
		UpdateLoginFlow(ctx).
		Flow(flow.Id).
		XSessionToken(token).
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
	logrus.Debugf("fleet: updated login flow: %v", successLogin)
	return &fleet.Session{
		Id:         successLogin.Session.Id,
		IdentityId: successLogin.Session.Identity.Id,
		Token:      *successLogin.SessionToken,
	}, nil
}

func (i *IdentityProvider) PerformAuthServerLoginFlowWithPasswordMethod(
	ctx context.Context,
	identifier, password string,
) (*fleet.Session, error) {
	flow, _, err := i.oryClient.FrontendApi.
		CreateNativeLoginFlow(ctx).
		Execute()
	if err != nil {
		return nil, err
	}
	logrus.Debugf("fleet: created native login flow: %v", flow)
	successLogin, _, err := i.oryClient.FrontendApi.
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
	logrus.Debugf("fleet: updated login flow: %v", successLogin)
	return &fleet.Session{
		Id:         successLogin.Session.Id,
		IdentityId: successLogin.Session.Identity.Id,
		Token:      *successLogin.SessionToken,
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
	logrus.Debugf("fleet: got session: %v", session)
	return &fleet.Session{
		Id:         session.Id,
		IdentityId: session.Identity.Id,
		Token:      token,
	}, nil
}

func (i *IdentityProvider) Logout(ctx context.Context, sessionToken string) error {
	logout, err := i.oryClient.FrontendApi.PerformNativeLogout(ctx).
		PerformNativeLogoutBody(
			*ory.NewPerformNativeLogoutBody(sessionToken),
		).
		Execute()
	if err != nil {
		return err
	}
	logrus.Debugf("fleet: logout: %v", logout)
	return nil
}
