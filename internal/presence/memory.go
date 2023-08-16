package presence

import (
	"context"
	"encoding/hex"
	"errors"
	"math/rand"

	fleet "go.shikanime.studio/elkia/pkg/api/fleet/v1alpha1"
)

type InMemoryIdentity struct {
	Username string
	Password string
}
type InMemoryIdentityServerConfig struct {
	Identities map[uint32]*InMemoryIdentity
	RandSource rand.Source
}

func NewInMemoryIdentityServer(cfg InMemoryIdentityServerConfig) *InMemoryIdentityServer {
	return &InMemoryIdentityServer{
		identities: cfg.Identities,
		rand:       rand.New(cfg.RandSource),
	}
}

type InMemoryIdentityServer struct {
	fleet.UnimplementedIdentityManagerServer
	identities map[uint32]*InMemoryIdentity
	rand       *rand.Rand
}

func (s *InMemoryIdentityServer) Login(
	ctx context.Context,
	in *fleet.LoginRequest,
) (*fleet.LoginResponse, error) {
	var identity *InMemoryIdentity
	for _, i := range s.identities {
		if i.Username == in.Identifier && i.Password == in.Password {
			identity = i
			break
		}
	}
	if identity == nil {
		return nil, errors.New("invalid credentials")
	}
	token, err := s.generateToken(16)
	if err != nil {
		return nil, err
	}
	return &fleet.LoginResponse{
		Token: token,
	}, nil
}

func (s *InMemoryIdentityServer) RefreshLogin(
	ctx context.Context,
	in *fleet.RefreshLoginRequest,
) (*fleet.RefreshLoginResponse, error) {
	_, err := s.Login(ctx, &fleet.LoginRequest{
		Identifier: in.Identifier,
		Password:   in.Password,
	})
	if err != nil {
		return nil, err
	}
	sessionToken, err := s.generateToken(16)
	if err != nil {
		return nil, err
	}
	return &fleet.RefreshLoginResponse{
		Token: sessionToken,
	}, nil
}

func (s *InMemoryIdentityServer) generateToken(length int) (string, error) {
	b := make([]byte, length)
	if _, err := s.rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

type InMemorySessionServerConfig struct {
	Sessions map[uint32]*fleet.Session
}

func NewInMemorySessionServer(cfg InMemorySessionServerConfig) *InMemorySessionServer {
	return &InMemorySessionServer{
		sessions: cfg.Sessions,
	}
}

type InMemorySessionServer struct {
	fleet.UnimplementedSessionManagerServer
	sessions map[uint32]*fleet.Session
}

func (s *InMemorySessionServer) GetSession(
	ctx context.Context,
	in *fleet.GetSessionRequest,
) (*fleet.GetSessionResponse, error) {
	session, ok := s.sessions[in.Code]
	if !ok {
		return nil, errors.New("session not found")
	}
	return &fleet.GetSessionResponse{
		Session: session,
	}, nil
}

func (s *InMemorySessionServer) PutSession(
	ctx context.Context,
	in *fleet.PutSessionRequest,
) (*fleet.PutSessionResponse, error) {
	s.sessions[in.Code] = in.Session
	return &fleet.PutSessionResponse{}, nil
}

func (s *InMemorySessionServer) DeleteSession(
	ctx context.Context,
	in *fleet.DeleteSessionRequest,
) (*fleet.DeleteSessionResponse, error) {
	delete(s.sessions, in.Code)
	return &fleet.DeleteSessionResponse{}, nil
}
