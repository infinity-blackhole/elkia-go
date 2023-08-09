package presence

import (
	"context"
	"encoding/gob"
	"fmt"
	"hash/fnv"

	ory "github.com/ory/client-go"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	fleet "go.shikanime.studio/elkia/pkg/api/fleet/v1alpha1"
	"google.golang.org/protobuf/proto"
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
	fleet.UnimplementedPresenceServer
	ory   *ory.APIClient
	redis redis.UniversalClient
}

func (i *PresenceServer) AuthCreateHandoffFlow(
	ctx context.Context,
	in *fleet.AuthCreateHandoffFlowRequest,
) (*fleet.AuthCreateHandoffFlowResponse, error) {
	login, err := i.AuthLogin(ctx, &fleet.AuthLoginRequest{
		Identifier: in.Identifier,
		Password:   in.Password,
	})
	if err != nil {
		return nil, err
	}
	sessionPut, err := i.SessionPut(ctx, &fleet.SessionPutRequest{
		Session: &fleet.Session{
			Token: login.Token,
		},
	})
	if err != nil {
		return nil, err
	}
	return &fleet.AuthCreateHandoffFlowResponse{
		Code: sessionPut.Code,
	}, nil
}

func (i *PresenceServer) AuthLogin(
	ctx context.Context,
	in *fleet.AuthLoginRequest,
) (*fleet.AuthLoginResponse, error) {
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
	return &fleet.AuthLoginResponse{
		Token: *successLogin.SessionToken,
	}, nil
}

func (s *PresenceServer) AuthRefreshLogin(
	ctx context.Context,
	in *fleet.AuthRefreshLoginRequest,
) (*fleet.AuthRefreshLoginResponse, error) {
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
	return &fleet.AuthRefreshLoginResponse{
		Token: *successLogin.SessionToken,
	}, nil
}

func (s *PresenceServer) AuthCompleteHandoffFlow(
	ctx context.Context,
	in *fleet.AuthCompleteHandoffFlowRequest,
) (*fleet.AuthCompleteHandoffFlowResponse, error) {
	sessionGet, err := s.SessionGet(ctx, &fleet.SessionGetRequest{
		Code: in.Code,
	})
	if err != nil {
		return nil, err
	}
	logrus.Debugf("fleet: got session: %v", sessionGet)
	refreshLogin, err := s.AuthRefreshLogin(
		ctx,
		&fleet.AuthRefreshLoginRequest{
			Identifier: in.Identifier,
			Password:   in.Password,
			Token:      sessionGet.Session.Token,
		},
	)
	if err != nil {
		return nil, err
	}
	_, err = s.SessionDelete(ctx, &fleet.SessionDeleteRequest{
		Code: in.Code,
	})
	if err != nil {
		return nil, err
	}
	logrus.Debugf("fleet: refreshed login")
	return &fleet.AuthCompleteHandoffFlowResponse{
		Token: refreshLogin.Token,
	}, nil
}

func (s *PresenceServer) AuthWhoAmI(
	ctx context.Context,
	in *fleet.AuthWhoAmIRequest,
) (*fleet.AuthWhoAmIResponse, error) {
	whoami, _, err := s.ory.FrontendApi.
		ToSession(ctx).
		Execute()
	if err != nil {
		return nil, err
	}
	logrus.Debugf("fleet: whoami: %v", whoami)
	return &fleet.AuthWhoAmIResponse{
		Id:         whoami.Id,
		IdentityId: whoami.Identity.Id,
	}, nil
}

func (s *PresenceServer) AuthLogout(
	ctx context.Context,
	in *fleet.AuthLogoutRequest,
) (*fleet.AuthLogoutResponse, error) {
	sessionGet, err := s.SessionGet(ctx, &fleet.SessionGetRequest{
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
	return &fleet.AuthLogoutResponse{}, nil
}

func (s *PresenceServer) SessionGet(
	ctx context.Context,
	in *fleet.SessionGetRequest,
) (*fleet.SessionGetResponse, error) {
	cmd := s.redis.Get(ctx, fmt.Sprintf("handoff_sessions:%d", in.Code))
	if err := cmd.Err(); err != nil {
		logrus.Tracef("fleet: got %s handoff sessions", err)
		return nil, err
	}
	res, err := cmd.Bytes()
	if err != nil {
		return nil, err
	}
	logrus.Debugf("fleet: got %s handoff sessions", res)
	var session fleet.Session
	if err := proto.Unmarshal(res, &session); err != nil {
		return nil, err
	}
	logrus.Debugf("fleet: got handoff session: %v", session.String())
	return &fleet.SessionGetResponse{
		Session: &session,
	}, nil
}

func (s *PresenceServer) SessionPut(
	ctx context.Context,
	in *fleet.SessionPutRequest,
) (*fleet.SessionPutResponse, error) {
	d, err := proto.Marshal(in.Session)
	if err != nil {
		return nil, err
	}
	code, err := s.generateCode(in.Session.Id)
	if err != nil {
		return nil, err
	}
	cmd := s.redis.Set(ctx, fmt.Sprintf("sessions:%d", code), d, 0)
	if err := cmd.Err(); err != nil {
		logrus.Tracef("fleet: got %s handoff sessions", err)
		return nil, err
	}
	logrus.Debugf("fleet: set handoff session: %d", code)
	return &fleet.SessionPutResponse{
		Code: code,
	}, nil
}

func (*PresenceServer) generateCode(id string) (uint32, error) {
	h := fnv.New32a()
	if err := gob.
		NewEncoder(h).
		Encode(id); err != nil {
		return 0, err
	}
	return h.Sum32(), nil
}

func (s *PresenceServer) SessionDelete(
	ctx context.Context,
	in *fleet.SessionDeleteRequest,
) (*fleet.SessionDeleteResponse, error) {
	cmd := s.redis.Del(ctx, fmt.Sprintf("sessions:%d", in.Code))
	if err := cmd.Err(); err != nil {
		logrus.Tracef("fleet: got %s handoff sessions", err)
		return nil, err
	}
	logrus.Debugf("fleet: deleted handoff session: %d", in.Code)
	return &fleet.SessionDeleteResponse{}, nil
}
