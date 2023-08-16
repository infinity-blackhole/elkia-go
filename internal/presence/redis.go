package presence

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	fleet "go.shikanime.studio/elkia/pkg/api/fleet/v1alpha1"
	"google.golang.org/protobuf/encoding/prototext"
)

type RedisSessionServerConfig struct {
	RedisClient redis.UniversalClient
}

func NewRedisSessionServer(cfg RedisSessionServerConfig) *RedisSessionServer {
	return &RedisSessionServer{
		redis: cfg.RedisClient,
	}
}

type RedisSessionServer struct {
	fleet.UnimplementedSessionManagerServer
	redis redis.UniversalClient
}

func (s *RedisSessionServer) GetSession(
	ctx context.Context,
	in *fleet.GetSessionRequest,
) (*fleet.GetSessionResponse, error) {
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
	var session fleet.Session
	if err := prototext.Unmarshal(res, &session); err != nil {
		return nil, err
	}
	logrus.Debugf("fleet: got handoff session: %v", session.String())
	return &fleet.GetSessionResponse{
		Session: &session,
	}, nil
}

func (s *RedisSessionServer) PutSession(
	ctx context.Context,
	in *fleet.PutSessionRequest,
) (*fleet.PutSessionResponse, error) {
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
	return &fleet.PutSessionResponse{}, nil
}

func (s *RedisSessionServer) DeleteSession(
	ctx context.Context,
	in *fleet.DeleteSessionRequest,
) (*fleet.DeleteSessionResponse, error) {
	cmd := s.redis.Del(ctx, fmt.Sprintf("sessions:%d", in.Code))
	if err := cmd.Err(); err != nil {
		logrus.Tracef("fleet: got %s handoff sessions", err)
		return nil, err
	}
	logrus.Debugf("fleet: deleted handoff session: %d", in.Code)
	return &fleet.DeleteSessionResponse{}, nil
}
