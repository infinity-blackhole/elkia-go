package gateway

import (
	"errors"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	eventing "go.shikanime.studio/elkia/pkg/api/eventing/v1alpha1"
	fleet "go.shikanime.studio/elkia/pkg/api/fleet/v1alpha1"
	world "go.shikanime.studio/elkia/pkg/api/world/v1alpha1"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type ServerConfig struct {
	PresenceClient fleet.PresenceClient
	LobbyClient    world.LobbyClient
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
	lobby    world.LobbyClient
	redis    redis.UniversalClient
}

func (s *Server) ChannelInteract(stream eventing.Gateway_ChannelInteractServer) error {
	logrus.Debug("gateway: channel interact")
	var sequence uint32
	msg, err := stream.Recv()
	if err != nil {
		return err
	}
	sync := msg.GetSync()
	if sync == nil {
		return errors.New("handoff: session protocol error")
	}
	sequence = sync.Sequence
	logrus.Debugf("gateway: channel interact: sync: %v", sync)
	msg, err = stream.Recv()
	if err != nil {
		return err
	}
	identifier := msg.GetIdentifier()
	if identifier.Sequence != sequence+1 {
		return errors.New("handoff: session protocol error")
	}
	sequence = identifier.Sequence
	logrus.Debugf("gateway: channel interact: identifier: %v", identifier)
	msg, err = stream.Recv()
	if err != nil {
		return err
	}
	password := msg.GetPassword()
	if password.Sequence != sequence+1 {
		return errors.New("handoff: session protocol error")
	}
	logrus.Debugf("gateway: channel interact: password: %v", password)
	handoff, err := s.presence.SubmitLoginFlow(stream.Context(), &fleet.SubmitLoginFlowRequest{
		Code:       sync.Code,
		Identifier: identifier.Value,
		Password:   password.Value,
	})
	if err != nil {
		logrus.Tracef("gateway: channel interact: handoff: %v", err)
		return err
	}
	logrus.Debugf("gateway: channel interact: handoff: %v", handoff)
	lobby, err := s.lobby.LobbyInteract(
		metadata.AppendToOutgoingContext(
			stream.Context(),
			"session",
			handoff.Token,
		),
	)
	if err != nil {
		logrus.Tracef("gateway: channel interact: lobby interact: %v", err)
		return err
	}
	r := Router{
		lobby: lobby,
	}
	return r.Serve(stream)
}

type Router struct {
	lobby world.Lobby_LobbyInteractClient
}

func (r *Router) Push(stream eventing.Gateway_ChannelInteractServer) error {
	for {
		msg, err := stream.Recv()
		if err != nil {
			return err
		}
		logrus.Tracef("gateway: controller proxy: publish: %s", msg)
		switch req := msg.Request.(type) {
		case *eventing.ChannelInteractRequest_ClientInteract:
		case *eventing.ChannelInteractRequest_LobbyInteract:
			if err := r.lobby.Send(req.LobbyInteract); err != nil {
				return err
			}
		default:
			return status.Newf(
				codes.Unimplemented,
				"unimplemented frame type: %T",
				msg.Request,
			).Err()
		}
	}
}

func (r *Router) Serve(stream eventing.Gateway_ChannelInteractServer) error {
	var wg errgroup.Group
	wg.Go(func() error {
		return r.Push(stream)
	})
	wg.Go(func() error {
		return r.Poll(stream)
	})
	return wg.Wait()
}

func (r *Router) Poll(stream eventing.Gateway_ChannelInteractServer) error {
	var wg errgroup.Group
	wg.Go(func() error {
		return r.PollLobby(stream)
	})
	return wg.Wait()
}

func (r *Router) PollLobby(stream eventing.Gateway_ChannelInteractServer) error {
	for {
		msg, err := r.lobby.Recv()
		if err != nil {
			return err
		}
		logrus.Tracef("gateway: controller proxy: poll: %s", msg)
		if err := stream.Send(&eventing.ChannelInteractResponse{
			Response: &eventing.ChannelInteractResponse_LobbyInteract{
				LobbyInteract: msg,
			},
		}); err != nil {
			return err
		}
	}
}
