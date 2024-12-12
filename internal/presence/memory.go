package presence

import (
	"context"
	"encoding/hex"
	"errors"
	"math/rand"

	fleet "go.shikanime.studio/elkia/pkg/api/fleet/v1alpha1"
)

type Identity struct {
	Username string
	Password string
}

type MemoryPresenceServerConfig struct {
	Identities map[uint32]*Identity
	Sessions   map[uint32]*fleet.Session
	Seed       int64
}

func NewMemoryPresenceServer(c MemoryPresenceServerConfig) *MemoryPresenceServer {
	return &MemoryPresenceServer{
		identities: c.Identities,
		sessions:   c.Sessions,
		rand:       rand.New(rand.NewSource(c.Seed)),
	}
}

type MemoryPresenceServer struct {
	fleet.UnimplementedPresenceServer
	identities map[uint32]*Identity
	sessions   map[uint32]*fleet.Session
	rand       *rand.Rand
}

func (s *MemoryPresenceServer) AuthCreateHandoffFlow(
	ctx context.Context,
	in *fleet.AuthCreateHandoffFlowRequest,
) (*fleet.AuthCreateHandoffFlowResponse, error) {
	var identity *Identity
	for _, i := range s.identities {
		if i.Username == in.Identifier && i.Password == in.Password {
			identity = i
			break
		}
	}
	if identity == nil {
		return nil, errors.New("invalid credentials")
	}
	sessionToken, err := s.generateSecureToken(16)
	if err != nil {
		return nil, err
	}
	code, err := generateCode(in.Identifier)
	if err != nil {
		return nil, err
	}
	sessionPut, err := s.SessionPut(ctx, &fleet.SessionPutRequest{
		Code: code,
		Session: &fleet.Session{
			Token: sessionToken,
		},
	})
	if err != nil {
		return nil, err
	}
	return &fleet.AuthCreateHandoffFlowResponse{
		Code: sessionPut.Code,
	}, nil
}

func (s *MemoryPresenceServer) AuthRefreshLogin(
	ctx context.Context,
	in *fleet.AuthRefreshHandoffFlowRequest,
) (*fleet.AuthRefreshHandoffFlowResponse, error) {
	var session *fleet.Session
	for _, i := range s.sessions {
		if i.Token == in.Token {
			session = i
			break
		}
	}
	if session == nil {
		return nil, errors.New("invalid session")
	}
	var identity *Identity
	for _, i := range s.identities {
		if i.Username == in.Identifier && i.Password == in.Password {
			identity = i
			break
		}
	}
	if identity == nil {
		return nil, errors.New("invalid credentials")
	}
	sessionToken, err := s.generateSecureToken(16)
	if err != nil {
		return nil, err
	}
	return &fleet.AuthRefreshHandoffFlowResponse{
		Token: sessionToken,
	}, nil
}

func (s *MemoryPresenceServer) AuthCompleteHandoffFlow(
	ctx context.Context,
	in *fleet.AuthCompleteHandoffFlowRequest,
) (*fleet.AuthCompleteHandoffFlowResponse, error) {
	code, err := generateCode(in.Identifier)
	if err != nil {
		return nil, err
	}
	sessionGet, err := s.SessionGet(ctx, &fleet.SessionGetRequest{
		Code: code,
	})
	if err != nil {
		return nil, err
	}
	_, err = s.AuthRefreshLogin(
		ctx,
		&fleet.AuthRefreshHandoffFlowRequest{
			Identifier: in.Identifier,
			Password:   in.Password,
			Token:      sessionGet.Session.Token,
		},
	)
	if err != nil {
		return nil, err
	}
	_, err = s.SessionDelete(ctx, &fleet.SessionDeleteRequest{
		Code: code,
	})
	if err != nil {
		return nil, err
	}
	return &fleet.AuthCompleteHandoffFlowResponse{}, nil
}

func (s *MemoryPresenceServer) generateSecureToken(length int) (string, error) {
	b := make([]byte, length)
	if _, err := s.rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func (s *MemoryPresenceServer) AuthLogout(
	ctx context.Context,
	in *fleet.AuthLogoutRequest,
) (*fleet.AuthLogoutResponse, error) {
	if _, err := s.SessionDelete(ctx, &fleet.SessionDeleteRequest{
		Code: in.Code,
	}); err != nil {
		return nil, err
	}
	return &fleet.AuthLogoutResponse{}, nil
}

func (s *MemoryPresenceServer) SessionGet(
	ctx context.Context,
	in *fleet.SessionGetRequest,
) (*fleet.SessionGetResponse, error) {
	session, ok := s.sessions[in.Code]
	if !ok {
		return nil, errors.New("session not found")
	}
	return &fleet.SessionGetResponse{
		Session: session,
	}, nil
}

func (s *MemoryPresenceServer) SessionPut(
	ctx context.Context,
	in *fleet.SessionPutRequest,
) (*fleet.SessionPutResponse, error) {
	s.sessions[in.Code] = in.Session
	return &fleet.SessionPutResponse{
		Code: in.Code,
	}, nil
}

func (s *MemoryPresenceServer) SessionDelete(
	ctx context.Context,
	in *fleet.SessionDeleteRequest,
) (*fleet.SessionDeleteResponse, error) {
	delete(s.sessions, in.Code)
	return &fleet.SessionDeleteResponse{}, nil
}
