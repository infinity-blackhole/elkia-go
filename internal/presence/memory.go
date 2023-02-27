package presence

import (
	"context"
	"encoding/gob"
	"encoding/hex"
	"errors"
	"hash/fnv"
	"math/rand"
	"strconv"

	fleet "github.com/infinity-blackhole/elkia/pkg/api/fleet/v1alpha1"
)

type Identity struct {
	Username string
	Password string
}

type MemoryPresenceServerConfig struct {
	Identities map[uint32]*Identity
	Seed       int64
}

func NewMemoryPresenceServer(c MemoryPresenceServerConfig) *MemoryPresenceServer {
	return &MemoryPresenceServer{
		identities: c.Identities,
		sessions:   map[uint32]*fleet.Session{},
		rand:       rand.New(rand.NewSource(c.Seed)),
	}
}

type MemoryPresenceServer struct {
	fleet.UnimplementedPresenceServer
	identities map[uint32]*Identity
	sessions   map[uint32]*fleet.Session
	rand       *rand.Rand
}

func (s *MemoryPresenceServer) AuthLogin(
	ctx context.Context,
	in *fleet.AuthLoginRequest,
) (*fleet.AuthLoginResponse, error) {
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
	sessionPut, err := s.SessionPut(ctx, &fleet.SessionPutRequest{
		Session: &fleet.Session{
			Id:    strconv.Itoa(s.rand.Int()),
			Token: sessionToken,
		},
	})
	if err != nil {
		return nil, err
	}
	return &fleet.AuthLoginResponse{
		Key: sessionPut.Key,
	}, nil
}

func (s *MemoryPresenceServer) AuthRefreshLogin(
	ctx context.Context,
	in *fleet.AuthRefreshLoginRequest,
) (*fleet.AuthRefreshLoginResponse, error) {
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
	sessionPut, err := s.SessionPut(ctx, &fleet.SessionPutRequest{
		Session: &fleet.Session{
			Id:    strconv.Itoa(s.rand.Int()),
			Token: sessionToken,
		},
	})
	if err != nil {
		return nil, err
	}
	return &fleet.AuthRefreshLoginResponse{
		Key: sessionPut.Key,
	}, nil
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
		Key: in.Key,
	}); err != nil {
		return nil, err
	}
	return &fleet.AuthLogoutResponse{}, nil
}

func (s *MemoryPresenceServer) SessionGet(
	ctx context.Context,
	in *fleet.SessionGetRequest,
) (*fleet.SessionGetResponse, error) {
	session, ok := s.sessions[in.Key]
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
	h := fnv.New32a()
	if err := gob.
		NewEncoder(h).
		Encode(in.Session.Id); err != nil {
		return nil, err
	}
	key := h.Sum32()
	s.sessions[key] = in.Session
	return &fleet.SessionPutResponse{
		Key: key,
	}, nil
}

func (s *MemoryPresenceServer) SessionDelete(
	ctx context.Context,
	in *fleet.SessionDeleteRequest,
) (*fleet.SessionDeleteResponse, error) {
	delete(s.sessions, in.Key)
	return &fleet.SessionDeleteResponse{}, nil
}
