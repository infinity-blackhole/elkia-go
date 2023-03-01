package presence

import (
	"context"
	"encoding/gob"
	"errors"
	"fmt"
	"hash/fnv"

	fleet "github.com/infinity-blackhole/elkia/pkg/api/fleet/v1alpha1"
	ory "github.com/ory/client-go"
	"github.com/sirupsen/logrus"
	etcd "go.etcd.io/etcd/client/v3"
	"google.golang.org/protobuf/proto"
)

type PresenceServerConfig struct {
	OryClient  *ory.APIClient
	EtcdClient *etcd.Client
}

func NewPresenceServer(config PresenceServerConfig) *PresenceServer {
	return &PresenceServer{
		ory:  config.OryClient,
		etcd: config.EtcdClient,
	}
}

type PresenceServer struct {
	fleet.UnimplementedPresenceServer
	ory  *ory.APIClient
	etcd *etcd.Client
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
	sessionPut, err := i.SessionPut(ctx, &fleet.SessionPutRequest{
		Session: &fleet.Session{
			Id:    successLogin.Session.Id,
			Token: *successLogin.SessionToken,
		},
	})
	if err != nil {
		return nil, err
	}
	return &fleet.AuthLoginResponse{
		Code: sessionPut.Code,
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
	sessionPut, err := s.SessionPut(ctx, &fleet.SessionPutRequest{
		Session: &fleet.Session{
			Id:    successLogin.Session.Id,
			Token: *successLogin.SessionToken,
		},
	})
	if err != nil {
		return nil, err
	}
	return &fleet.AuthRefreshLoginResponse{
		Code: sessionPut.Code,
	}, nil
}

func (s *PresenceServer) AuthHandoff(
	ctx context.Context,
	in *fleet.AuthHandoffRequest,
) (*fleet.AuthHandoffResponse, error) {
	sessionGet, err := s.SessionGet(ctx, &fleet.SessionGetRequest{
		Code: in.Code,
	})
	if err != nil {
		return nil, err
	}
	logrus.Debugf("fleet: got session: %v", sessionGet)
	_, err = s.AuthRefreshLogin(
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
	logrus.Debugf("fleet: refreshed login")
	return &fleet.AuthHandoffResponse{}, nil
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
	res, err := s.etcd.Get(ctx, fmt.Sprintf("handoff_sessions:%d", in.Code))
	if err != nil {
		return nil, err
	}
	logrus.Debugf("fleet: got %d handoff sessions", len(res.Kvs))
	if len(res.Kvs) == 1 {
		return nil, errors.New("invalid code")
	}
	var session fleet.Session
	if err := proto.Unmarshal(res.Kvs[0].Value, &session); err != nil {
		return nil, err
	}
	logrus.Debugf("fleet: got handoff session: %v", session)
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
	res, err := s.etcd.Put(ctx, fmt.Sprintf("sessions:%d", code), string(d))
	if err != nil {
		return nil, err
	}
	logrus.Debugf("fleet: set handoff session: %v", res)
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
	res, err := s.etcd.Delete(ctx, fmt.Sprintf("sessions:%d", in.Code))
	if err != nil {
		return nil, err
	}
	logrus.Debugf("fleet: deleted handoff session: %v", res)
	return &fleet.SessionDeleteResponse{}, nil
}
