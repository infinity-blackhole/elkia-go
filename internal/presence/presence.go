package presence

import (
	"context"
	"fmt"

	ory "github.com/ory/client-go"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	fleetpb "go.shikanime.studio/elkia/pkg/api/fleet/v1alpha1"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/types/known/emptypb"
)

type PresenceServerConfig struct {
	OryClient   *ory.APIClient
	RedisClient redis.UniversalClient
}

func NewPresenceServer(config PresenceServerConfig) *PresenceServer {
	return &PresenceServer{
		ory:   config.OryClient,
		redis: config.RedisClient,
	}
}

type PresenceServer struct {
	fleetpb.UnimplementedPresenceServer
	ory   *ory.APIClient
	redis redis.UniversalClient
}

func (s *PresenceServer) CreateHandoffFlow(
	ctx context.Context,
	in *fleetpb.CreateHandoffFlowRequest,
) (*fleetpb.CreateHandoffFlowResponse, error) {
	login, err := s.Login(ctx, &fleetpb.LoginRequest{
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
	sessionPut, err := s.PutSession(ctx, &fleetpb.PutSessionRequest{
		Code: code,
		Session: &fleetpb.Session{
			Token: login.Token,
		},
	})
	if err != nil {
		return nil, err
	}
	return &fleetpb.CreateHandoffFlowResponse{
		Code: sessionPut.Code,
	}, nil
}

func (i *PresenceServer) Login(
	ctx context.Context,
	in *fleetpb.LoginRequest,
) (*fleetpb.LoginResponse, error) {
	flow, _, err := i.ory.FrontendApi.
		CreateNativeLoginFlow(ctx).
		Execute()
	if err != nil {
		return nil, err
	}
	logrus.Debugf("fleet: created native login flow: %v", flow)
	successLogin, _, err := i.ory.FrontendApi.
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
	return &fleetpb.LoginResponse{
		Token: *successLogin.SessionToken,
	}, nil
}

func (s *PresenceServer) RefreshLogin(
	ctx context.Context,
	in *fleetpb.RefreshLoginRequest,
) (*fleetpb.RefreshLoginResponse, error) {
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
	return &fleetpb.RefreshLoginResponse{
		Token: *successLogin.SessionToken,
	}, nil
}

func (s *PresenceServer) CompleteHandoffFlow(
	ctx context.Context,
	in *fleetpb.CompleteHandoffFlowRequest,
) (*fleetpb.CompleteHandoffFlowResponse, error) {
	code, err := generateCode(in.Identifier)
	if err != nil {
		return nil, err
	}
	sessionGet, err := s.GetSession(ctx, &fleetpb.GetSessionRequest{
		Code: code,
	})
	if err != nil {
		return nil, err
	}
	logrus.Debugf("fleet: got session: %v", sessionGet)
	refreshLogin, err := s.RefreshLogin(
		ctx,
		&fleetpb.RefreshLoginRequest{
			Identifier: in.Identifier,
			Password:   in.Password,
			Token:      sessionGet.Session.Token,
		},
	)
	if err != nil {
		return nil, err
	}
	_, err = s.DeleteSession(ctx, &fleetpb.DeleteSessionRequest{
		Code: code,
	})
	if err != nil {
		return nil, err
	}
	logrus.Debugf("fleet: refreshed login")
	return &fleetpb.CompleteHandoffFlowResponse{
		Token: refreshLogin.Token,
	}, nil
}

func (s *PresenceServer) WhoAmI(
	ctx context.Context,
	in *fleetpb.WhoAmIRequest,
) (*fleetpb.WhoAmIResponse, error) {
	whoami, _, err := s.ory.FrontendApi.
		ToSession(ctx).
		XSessionToken(in.Token).
		Execute()
	if err != nil {
		return nil, err
	}
	logrus.Debugf("fleet: whoami: %v", whoami)
	return &fleetpb.WhoAmIResponse{
		Id:         whoami.Id,
		IdentityId: whoami.Identity.Id,
	}, nil
}

func (s *PresenceServer) Logout(
	ctx context.Context,
	in *fleetpb.LogoutRequest,
) (*emptypb.Empty, error) {
	sessionGet, err := s.GetSession(ctx, &fleetpb.GetSessionRequest{
		Code: in.Code,
	})
	if err != nil {
		return nil, err
	}
	logrus.Debugf("fleet: got session: %v", sessionGet)
	_, err = s.ory.FrontendApi.
		PerformNativeLogout(ctx).
		PerformNativeLogoutBody(
			*ory.NewPerformNativeLogoutBody(sessionGet.Session.Token),
		).
		Execute()
	if err != nil {
		return nil, err
	}
	logrus.Debugf("fleet: revoked session")
	return &emptypb.Empty{}, nil
}

func (s *PresenceServer) GetSession(
	ctx context.Context,
	in *fleetpb.GetSessionRequest,
) (*fleetpb.GetSessionResponse, error) {
	cmd := s.redis.Get(ctx, fmt.Sprintf("sessions:%d", in.Code))
	if err := cmd.Err(); err != nil {
		logrus.Tracef("fleet: got %s handoff sessions", err)
		return nil, err
	}
	res, err := cmd.Bytes()
	if err != nil {
		return nil, err
	}
	logrus.Debugf("fleet: got %s handoff sessions", res)
	var session fleetpb.Session
	if err := prototext.Unmarshal(res, &session); err != nil {
		return nil, err
	}
	logrus.Debugf("fleet: got handoff session: %v", session.String())
	return &fleetpb.GetSessionResponse{
		Session: &session,
	}, nil
}

func (s *PresenceServer) PutSession(
	ctx context.Context,
	in *fleetpb.PutSessionRequest,
) (*fleetpb.PutSessionRequest, error) {
	d, err := prototext.Marshal(in.Session)
	if err != nil {
		return nil, err
	}
	cmd := s.redis.Set(ctx, fmt.Sprintf("sessions:%d", in.Code), d, 0)
	if err := cmd.Err(); err != nil {
		logrus.Tracef("fleet: got %s handoff sessions", err)
		return nil, err
	}
	logrus.Debugf("fleet: set handoff session: %d", in.Code)
	return &fleetpb.PutSessionRequest{}, nil
}

func (s *PresenceServer) DeleteSession(
	ctx context.Context,
	in *fleetpb.DeleteSessionRequest,
) (*emptypb.Empty, error) {
	cmd := s.redis.Del(ctx, fmt.Sprintf("sessions:%d", in.Code))
	if err := cmd.Err(); err != nil {
		logrus.Tracef("fleet: got %s handoff sessions", err)
		return nil, err
	}
	logrus.Debugf("fleet: deleted handoff session: %d", in.Code)
	return &emptypb.Empty{}, nil
}
