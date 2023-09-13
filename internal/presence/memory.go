package presence

import (
	"context"
	"encoding/hex"
	"errors"
	"math/rand"

	fleetpb "go.shikanime.studio/elkia/pkg/api/fleet/v1alpha1"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Identity struct {
	Username string
	Password string
}

type MemoryPresenceServerConfig struct {
	Identities map[uint32]*Identity
	Sessions   map[uint32]*fleetpb.Session
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
	fleetpb.UnimplementedPresenceServer
	identities map[uint32]*Identity
	sessions   map[uint32]*fleetpb.Session
	rand       *rand.Rand
}

func (s *MemoryPresenceServer) CreateHandoffFlow(
	ctx context.Context,
	in *fleetpb.CreateHandoffFlowRequest,
) (*fleetpb.CreateHandoffFlowResponse, error) {
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
	sessionPut, err := s.PutSession(ctx, &fleetpb.PutSessionRequest{
		Code: code,
		Session: &fleetpb.Session{
			Token: sessionToken,
		},
	})
	if err != nil {
		return nil, err
	}
	return &fleetpb.CreateHandoffFlowResponse{
		Code: sessionPut.Code,
	}, nil
}

func (s *MemoryPresenceServer) RefreshLogin(
	ctx context.Context,
	in *fleetpb.RefreshLoginRequest,
) (*fleetpb.RefreshLoginResponse, error) {
	var session *fleetpb.Session
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
	return &fleetpb.RefreshLoginResponse{
		Token: sessionToken,
	}, nil
}

func (s *MemoryPresenceServer) CompleteHandoffFlow(
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
	_, err = s.RefreshLogin(
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
	return &fleetpb.CompleteHandoffFlowResponse{}, nil
}

func (s *MemoryPresenceServer) generateSecureToken(length int) (string, error) {
	b := make([]byte, length)
	if _, err := s.rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func (s *MemoryPresenceServer) Logout(
	ctx context.Context,
	in *fleetpb.LogoutRequest,
) (*emptypb.Empty, error) {
	if _, err := s.DeleteSession(ctx, &fleetpb.DeleteSessionRequest{
		Code: in.Code,
	}); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s *MemoryPresenceServer) GetSession(
	ctx context.Context,
	in *fleetpb.GetSessionRequest,
) (*fleetpb.GetSessionResponse, error) {
	session, ok := s.sessions[in.Code]
	if !ok {
		return nil, errors.New("session not found")
	}
	return &fleetpb.GetSessionResponse{
		Session: session,
	}, nil
}

func (s *MemoryPresenceServer) PutSession(
	ctx context.Context,
	in *fleetpb.PutSessionRequest,
) (*fleetpb.PutSessionResponse, error) {
	s.sessions[in.Code] = in.Session
	return &fleetpb.PutSessionResponse{
		Code: in.Code,
	}, nil
}

func (s *MemoryPresenceServer) DeleteSession(
	ctx context.Context,
	in *fleetpb.DeleteSessionRequest,
) (*emptypb.Empty, error) {
	delete(s.sessions, in.Code)
	return &emptypb.Empty{}, nil
}
