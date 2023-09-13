package gateway

import (
	"errors"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	eventingpb "go.shikanime.studio/elkia/pkg/api/eventing/v1alpha1"
	fleetpb "go.shikanime.studio/elkia/pkg/api/fleet/v1alpha1"
	worldpb "go.shikanime.studio/elkia/pkg/api/world/v1alpha1"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/encoding/prototext"
)

type ServerConfig struct {
	PresenceClient fleetpb.PresenceClient
	LobbyClient    worldpb.LobbyClient
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
	eventingpb.UnimplementedGatewayServer
	presence fleetpb.PresenceClient
	lobby    worldpb.LobbyClient
	redis    redis.UniversalClient
}

func (s *Server) GatewayInteract(stream eventingpb.Gateway_ChannelInteractServer) error {
	logrus.Debug("gateway: channel interact")
	var sequence uint32
	msg, err := stream.Recv()
	if err != nil {
		return err
	}
	sync := msg.GetSyncCommand()
	if sync == nil {
		return errors.New("handoff: session protocol error")
	}
	sequence = sync.Sequence
	logrus.Debugf("gateway: channel interact: sync: %v", sync)
	msg, err = stream.Recv()
	if err != nil {
		return err
	}
	identifier := msg.GetIdentifierCommand()
	if identifier.Sequence != sequence+1 {
		return errors.New("handoff: session protocol error")
	}
	sequence = identifier.Sequence
	logrus.Debugf("gateway: channel interact: identifier: %v", identifier)
	msg, err = stream.Recv()
	if err != nil {
		return err
	}
	password := msg.GetPasswordCommand()
	if password.Sequence != sequence+1 {
		return errors.New("handoff: session protocol error")
	}
	logrus.Debugf("gateway: channel interact: password: %v", password)
	handoff, err := s.presence.CompleteHandoffFlow(stream.Context(), &fleetpb.CompleteHandoffFlowRequest{
		Identifier: identifier.Identifier,
		Password:   password.Password,
	})
	if err != nil {
		logrus.Tracef("gateway: channel interact: handoff: %v", err)
		return err
	}
	logrus.Debugf("gateway: channel interact: handoff: %v", handoff)
	ctr := &ControllerProxy{
		redis:    s.redis,
		presence: s.presence,
		token:    handoff.Token,
	}
	return ctr.Serve(stream)
}

type ControllerProxy struct {
	redis    redis.UniversalClient
	presence fleetpb.PresenceClient
	token    string
}

func (s *ControllerProxy) Push(stream eventingpb.Gateway_ChannelInteractServer) error {
	whoami, err := s.presence.WhoAmI(stream.Context(), &fleetpb.WhoAmIRequest{
		Token: s.token,
	})
	if err != nil {
		return err
	}
	logrus.Debugf("gateway: controller proxy: whoami: %v", whoami)
	for {
		msg, err := stream.Recv()
		if err != nil {
			return err
		}
		data, err := prototext.Marshal(msg)
		if err != nil {
			return err
		}
		logrus.Tracef("gateway: controller proxy: publish: %s", data)
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

func (s *ControllerProxy) Poll(stream eventingpb.Gateway_ChannelInteractServer) error {
	whoami, err := s.presence.WhoAmI(stream.Context(), &fleetpb.WhoAmIRequest{
		Token: s.token,
	})
	if err != nil {
		return err
	}
	pubsub := s.redis.Subscribe(stream.Context(), fmt.Sprintf("elkia:controllers:events:%s", whoami.IdentityId))
	for msg := range pubsub.Channel() {
		var frame eventingpb.GatewayEvent
		if err := prototext.Unmarshal([]byte(msg.Payload), &frame); err != nil {
			return err
		}
		logrus.Tracef("gateway: controller proxy: publish: %v", frame)
		if err := stream.Send(&frame); err != nil {
			return err
		}
	}
	return nil
}

func (s *ControllerProxy) Serve(stream eventingpb.Gateway_ChannelInteractServer) error {
	var wg errgroup.Group
	wg.Go(func() error {
		return s.Push(stream)
	})
	wg.Go(func() error {
		return s.Poll(stream)
	})
	return wg.Wait()
}
