package gateway

import (
	"errors"
	"fmt"

	"github.com/redis/go-redis/v9"
	eventing "go.shikanime.studio/elkia/pkg/api/eventing/v1alpha1"
	fleet "go.shikanime.studio/elkia/pkg/api/fleet/v1alpha1"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/proto"
)

type ServerConfig struct {
	PresenceClient fleet.PresenceClient
	LobbyClient    eventing.LobbyClient
	RedisClient    redis.UniversalClient
}

func NewServer(cfg ServerConfig) *Server {
	return &Server{
		presence: cfg.PresenceClient,
		lobby:    cfg.LobbyClient,
		redis:    cfg.RedisClient,
	}
}

type Server struct {
	eventing.UnimplementedGatewayServer
	presence fleet.PresenceClient
	lobby    eventing.LobbyClient
	redis    redis.UniversalClient
}

func (s *Server) ChannelInteract(stream eventing.Gateway_ChannelInteractServer) error {
	var sequence uint32
	msg, err := stream.Recv()
	if err != nil {
		return err
	}
	sync := msg.GetSyncFrame()
	if sync == nil {
		return errors.New("handoff: session protocol error")
	}
	sequence = sync.Sequence
	msg, err = stream.Recv()
	if err != nil {
		return err
	}
	identifier := msg.GetIdentifierFrame()
	if identifier.Sequence != sequence+1 {
		return errors.New("handoff: session protocol error")
	}
	sequence = identifier.Sequence
	msg, err = stream.Recv()
	if err != nil {
		return err
	}
	password := msg.GetPasswordFrame()
	if password.Sequence != sequence+1 {
		return errors.New("handoff: session protocol error")
	}
	handoff, err := s.presence.AuthCompleteHandoffFlow(stream.Context(), &fleet.AuthCompleteHandoffFlowRequest{
		Code:       sync.Code,
		Identifier: identifier.Identifier,
		Password:   password.Password,
	})
	if err != nil {
		return err
	}
	return NewControllerProxy(s.redis, handoff.Token).Serve(stream)
}

func NewControllerProxy(redis redis.UniversalClient, token string) *ControllerProxy {
	return &ControllerProxy{
		redis: redis,
		token: token,
	}
}

type ControllerProxy struct {
	redis    redis.UniversalClient
	presence fleet.PresenceClient
	token    string
}

func (s *ControllerProxy) Push(stream eventing.Gateway_ChannelInteractServer) error {
	whoami, err := s.presence.AuthWhoAmI(stream.Context(), &fleet.AuthWhoAmIRequest{
		Token: s.token,
	})
	if err != nil {
		return err
	}
	for {
		msg, err := stream.Recv()
		if err != nil {
			return err
		}
		data, err := proto.Marshal(msg)
		if err != nil {
			return err
		}
		cmdRes := s.redis.Publish(
			stream.Context(),
			fmt.Sprintf("elkia:controllers:commands:%s", whoami.IdentityId),
			data,
		)
		if err := cmdRes.Err(); err != nil {
			return err
		}
	}
}

func (s *ControllerProxy) Poll(stream eventing.Gateway_ChannelInteractServer) error {
	whoami, err := s.presence.AuthWhoAmI(stream.Context(), &fleet.AuthWhoAmIRequest{
		Token: s.token,
	})
	if err != nil {
		return err
	}
	pubsub := s.redis.Subscribe(stream.Context(), fmt.Sprintf("elkia:controllers:events:%s", whoami.IdentityId))
	for msg := range pubsub.Channel() {
		var frame eventing.ChannelInteractResponse
		if err := proto.Unmarshal([]byte(msg.Payload), &frame); err != nil {
			return err
		}
		if err := stream.Send(&frame); err != nil {
			return err
		}
	}
	return nil
}

func (s *ControllerProxy) Serve(stream eventing.Gateway_ChannelInteractServer) error {
	wg := errgroup.Group{}
	wg.Go(func() error {
		return s.Push(stream)
	})
	wg.Go(func() error {
		return s.Poll(stream)
	})
	return wg.Wait()
}
