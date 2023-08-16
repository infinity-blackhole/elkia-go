package presence

import (
	"context"

	"github.com/sirupsen/logrus"
	eventing "go.shikanime.studio/elkia/pkg/api/eventing/v1alpha1"
	fleet "go.shikanime.studio/elkia/pkg/api/fleet/v1alpha1"
	"go.shikanime.studio/elkia/pkg/protonostale"
)

type PresenceServerConfig struct {
	SessionManager  fleet.SessionManagerServer
	IdentityManager fleet.IdentityManagerServer
}

func NewPresenceServer(cfg PresenceServerConfig) *PresenceServer {
	return &PresenceServer{
		identity: cfg.IdentityManager,
		session:  cfg.SessionManager,
	}
}

type PresenceServer struct {
	fleet.UnimplementedPresenceServer
	identity fleet.IdentityManagerServer
	session  fleet.SessionManagerServer
}

func (s *PresenceServer) CreateLoginFlow(
	ctx context.Context,
	in *fleet.CreateLoginFlowRequest,
) (*fleet.CreateLoginFlowResponse, error) {
	login, err := s.identity.Login(ctx, &fleet.LoginRequest{
		Identifier: in.Identifier,
		Password:   in.Password,
	})
	if err != nil {
		return nil, err
	}
	code, err := generateCode(in.Identifier)
	if err != nil {
		return nil, err
	}
	_, err = s.session.PutSession(ctx, &fleet.PutSessionRequest{
		Code: code,
		Session: &fleet.Session{
			State: &fleet.Session_Handoff{
				Handoff: &fleet.SessionHandoff{
					Token: login.Token,
				},
			},
		},
	})
	if err != nil {
		return nil, err
	}
	return &fleet.CreateLoginFlowResponse{
		Code: code,
	}, nil
}

func (s *PresenceServer) SubmitLoginFlow(
	ctx context.Context,
	in *fleet.SubmitLoginFlowRequest,
) (*fleet.SubmitLoginFlowResponse, error) {
	code, err := generateCode(in.Identifier)
	if err != nil {
		return nil, err
	}
	sessionGet, err := s.session.GetSession(ctx, &fleet.GetSessionRequest{
		Code: code,
	})
	if err != nil {
		return nil, err
	}
	logrus.Debugf("fleet: got session: %v", sessionGet)
	switch state := sessionGet.Session.State.(type) {
	case *fleet.Session_Handoff:
		claim, err := s.ClaimLoginFlow(ctx, &fleet.ClaimLoginFlowRequest{
			Identifier: in.Identifier,
			Password:   in.Password,
			Token:      state.Handoff.Token,
		})
		return &fleet.SubmitLoginFlowResponse{
			Token: claim.Token,
		}, err
	case *fleet.Session_Online:
		return nil, protonostale.NewStatus(eventing.Code_SESSION_ALREADY_USED)
	default:
		return nil, protonostale.NewStatus(eventing.Code_UNEXPECTED_ERROR)
	}
}

func (s *PresenceServer) ClaimLoginFlow(
	ctx context.Context,
	in *fleet.ClaimLoginFlowRequest,
) (*fleet.ClaimLoginFlowResponse, error) {
	code, err := generateCode(in.Identifier)
	if err != nil {
		return nil, err
	}
	refreshLogin, err := s.identity.RefreshLogin(
		ctx,
		&fleet.RefreshLoginRequest{
			Identifier: in.Identifier,
			Password:   in.Password,
			Token:      in.Token,
		},
	)
	if err != nil {
		return nil, err
	}
	_, err = s.session.PutSession(ctx, &fleet.PutSessionRequest{
		Code: code,
		Session: &fleet.Session{
			State: &fleet.Session_Online{
				Online: &fleet.SessionOnline{},
			},
		},
	})
	if err != nil {
		return nil, err
	}
	logrus.Debugf("fleet: refreshed login")
	return &fleet.ClaimLoginFlowResponse{
		Token: refreshLogin.Token,
	}, nil
}

func (s *PresenceServer) Logout(
	ctx context.Context,
	in *fleet.LogoutRequest,
) (*fleet.LogoutResponse, error) {
	whoami, err := s.identity.WhoAmI(ctx, &fleet.WhoAmIRequest{
		Token: in.Token,
	})
	if err != nil {
		return nil, err
	}
	logrus.Debugf("fleet: whoami: %v", whoami)
	_, err = s.identity.Logout(ctx, &fleet.LogoutRequest{
		Token: in.Token,
	})
	if err != nil {
		return nil, err
	}
	logrus.Debugf("fleet: revoked login")
	code, err := generateCode(whoami.IdentityId)
	if err != nil {
		return nil, err
	}
	_, err = s.session.DeleteSession(ctx, &fleet.DeleteSessionRequest{
		Code: code,
	})
	if err != nil {
		return nil, err
	}
	return &fleet.LogoutResponse{}, nil
}
