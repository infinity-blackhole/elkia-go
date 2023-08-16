package presence

import (
	"context"

	ory "github.com/ory/client-go"
	"github.com/sirupsen/logrus"
	fleet "go.shikanime.studio/elkia/pkg/api/fleet/v1alpha1"
)

type OryIdentityServerConfig struct {
	OryClient *ory.APIClient
}

func NewOryIdentityServer(cfg OryIdentityServerConfig) *OryIdentityServer {
	return &OryIdentityServer{
		ory: cfg.OryClient,
	}
}

type OryIdentityServer struct {
	fleet.UnimplementedIdentityManagerServer
	ory *ory.APIClient
}

func (s *OryIdentityServer) Login(
	ctx context.Context,
	in *fleet.LoginRequest,
) (*fleet.LoginResponse, error) {
	flow, _, err := s.ory.FrontendApi.
		CreateNativeLoginFlow(ctx).
		Execute()
	if err != nil {
		return nil, err
	}
	logrus.Debugf("fleet: created native login flow: %v", flow)
	successLogin, _, err := s.ory.FrontendApi.
		UpdateLoginFlow(ctx).
		Flow(flow.Id).
		UpdateLoginFlowBody(
			ory.UpdateLoginFlowWithPasswordMethodAsUpdateLoginFlowBody(
				ory.NewUpdateLoginFlowWithPasswordMethod(
					in.Identifier,
					"password",
					in.Password,
				),
			),
		).
		Execute()
	if err != nil {
		return nil, err
	}
	logrus.Debugf("fleet: updated login flow: %v", successLogin)
	return &fleet.LoginResponse{
		Token: *successLogin.SessionToken,
	}, nil
}

func (s *OryIdentityServer) RefreshLogin(
	ctx context.Context,
	in *fleet.RefreshLoginRequest,
) (*fleet.RefreshLoginResponse, error) {
	flow, _, err := s.ory.FrontendApi.
		CreateNativeLoginFlow(ctx).
		XSessionToken(in.Token).
		Refresh(true).
		Execute()
	if err != nil {
		return nil, err
	}
	logrus.Debugf("fleet: created native login flow: %v", flow)
	successLogin, _, err := s.ory.FrontendApi.
		UpdateLoginFlow(ctx).
		Flow(flow.Id).
		XSessionToken(in.Token).
		UpdateLoginFlowBody(
			ory.UpdateLoginFlowWithPasswordMethodAsUpdateLoginFlowBody(
				ory.NewUpdateLoginFlowWithPasswordMethod(
					in.Identifier,
					"password",
					in.Password,
				),
			),
		).
		Execute()
	if err != nil {
		return nil, err
	}
	logrus.Debugf("fleet: updated login flow: %v", successLogin)
	return &fleet.RefreshLoginResponse{
		Token: *successLogin.SessionToken,
	}, nil
}

func (s *OryIdentityServer) Logout(
	ctx context.Context,
	in *fleet.LogoutRequest,
) (*fleet.LogoutResponse, error) {
	_, err := s.ory.FrontendApi.
		PerformNativeLogout(ctx).
		PerformNativeLogoutBody(
			*ory.NewPerformNativeLogoutBody(in.Token),
		).
		Execute()
	if err != nil {
		return nil, err
	}
	logrus.Debugf("fleet: revoked session")
	return &fleet.LogoutResponse{}, nil
}

func (s *OryIdentityServer) WhoAmI(
	ctx context.Context,
	in *fleet.WhoAmIRequest,
) (*fleet.WhoAmIResponse, error) {
	whoami, _, err := s.ory.FrontendApi.
		ToSession(ctx).
		XSessionToken(in.Token).
		Execute()
	if err != nil {
		return nil, err
	}
	logrus.Debugf("fleet: whoami: %v", whoami)
	return &fleet.WhoAmIResponse{
		Id:         whoami.Id,
		IdentityId: whoami.Identity.Id,
	}, nil
}
