package gateway

import (
	"errors"

	eventing "github.com/infinity-blackhole/elkia/pkg/api/eventing/v1alpha1"
	fleet "github.com/infinity-blackhole/elkia/pkg/api/fleet/v1alpha1"
	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/proto"
)

type ServerConfig struct {
	PresenceClient fleet.PresenceClient
	RedisClient    redis.UniversalClient
}

func NewServer(cfg ServerConfig) *Server {
	return &Server{
		presence: cfg.PresenceClient,
		redis:    cfg.RedisClient,
	}
}

type Server struct {
	eventing.UnimplementedGatewayServer
	presence fleet.PresenceClient
	redis    redis.UniversalClient
	sequence uint32
}

func (s *Server) ChannelInteract(stream eventing.Gateway_ChannelInteractServer) error {
	msg, err := stream.Recv()
	if err != nil {
		return err
	}
	sync := msg.GetSyncFrame()
	if sync == nil {
		return errors.New("handoff: session protocol error")
	}
	s.sequence = sync.Sequence
	msg, err = stream.Recv()
	if err != nil {
		return err
	}
	identifier := msg.GetIdentifierFrame()
	if identifier.Sequence != s.sequence+1 {
		return errors.New("handoff: session protocol error")
	}
	s.sequence = identifier.Sequence
	msg, err = stream.Recv()
	if err != nil {
		return err
	}
	password := msg.GetPasswordFrame()
	if password.Sequence != s.sequence+1 {
		return errors.New("handoff: session protocol error")
	}
	s.sequence = password.Sequence
	_, err = s.presence.AuthHandoff(stream.Context(), &fleet.AuthHandoffRequest{
		Code:       sync.Code,
		Identifier: identifier.Identifier,
		Password:   password.Password,
	})
	if err != nil {
		return err
	}
	wg := errgroup.Group{}
	wg.Go(func() error {
		return s.Push(stream)
	})
	wg.Go(func() error {
		return s.Poll(stream)
	})
	return wg.Wait()
}

func (s *Server) Push(stream eventing.Gateway_ChannelInteractServer) error {
	for {
		msg, err := stream.Recv()
		if err != nil {
			return err
		}
		switch msg.Payload.(type) {
		case *eventing.ChannelInteractRequest_CommandFrame:
			command := msg.GetCommandFrame()
			switch command.Payload.(type) {
			case *eventing.CommandFrame_HeartbeatFrame:
				if command.Sequence != s.sequence+1 {
					return errors.New("handoff: channel protocol error")
				}
				s.sequence = command.Sequence
			default:
				cmdRes := s.redis.Publish(
					stream.Context(),
					"elkia:channel:command",
					command,
				)
				if err := cmdRes.Err(); err != nil {
					return err
				}
			}
		default:
			return errors.New("channel: channel protocol error")
		}
	}
}

func (s *Server) Poll(stream eventing.Gateway_ChannelInteractServer) error {
	pubsub := s.redis.Subscribe(stream.Context(), "elkia:channel:response")
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
