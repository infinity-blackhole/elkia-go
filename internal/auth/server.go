package auth

import (
	"math"
	"net"

	"github.com/sirupsen/logrus"
	eventingpb "go.shikanime.studio/elkia/pkg/api/eventing/v1alpha1"
	fleetpb "go.shikanime.studio/elkia/pkg/api/fleet/v1alpha1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ServerConfig struct {
	PresenceClient fleetpb.PresenceClient
	ClusterClient  fleetpb.ClusterClient
}

func NewServer(cfg ServerConfig) *Server {
	return &Server{
		presence: cfg.PresenceClient,
		cluster:  cfg.ClusterClient,
	}
}

type Server struct {
	eventingpb.UnimplementedAuthServer
	presence fleetpb.PresenceClient
	cluster  fleetpb.ClusterClient
}

func (s *Server) AuthInteract(stream eventingpb.Auth_AuthInteractServer) error {
	for {
		m, err := stream.Recv()
		if err != nil {
			return err
		}
		switch m.Command.(type) {
		case *eventingpb.AuthCommand_CreateHandoffFlow:
			logrus.Debugf("auth: handle handoff")
			if err := s.HandleCreateHandoffFlowCommand(
				m.GetCreateHandoffFlow(),
				stream,
			); err != nil {
				return err
			}
		default:
			return status.Newf(codes.Unimplemented, "unimplemented frame type: %T", m.Command).Err()
		}
	}
}

func (s *Server) HandleCreateHandoffFlowCommand(
	m *eventingpb.CreateHandoffFlowCommand,
	stream eventingpb.Auth_AuthInteractServer,
) error {
	handoff, err := s.presence.CreateHandoffFlow(
		stream.Context(),
		&fleetpb.CreateHandoffFlowRequest{
			Identifier: m.Identifier,
			Password:   m.Password,
		},
	)
	if err != nil {
		return err
	}
	logrus.Debugf("auth: create handoff: %v", handoff)

	memberList, err := s.cluster.MemberList(
		stream.Context(),
		&fleetpb.MemberListRequest{},
	)
	if err != nil {
		return err
	}
	logrus.Debugf("auth: list members: %v", memberList)
	var ms []*fleetpb.Endpoint
	for _, m := range memberList.Members {
		for _, a := range m.Addresses {
			host, port, err := net.SplitHostPort(a)
			if err != nil {
				return err
			}
			ms = append(ms, &fleetpb.Endpoint{
				Host:      host,
				Port:      port,
				Weight:    uint32(math.Round(float64(m.Population)/float64(m.Capacity)*20) + 1),
				WorldId:   m.WorldId,
				ChannelId: m.ChannelId,
				WorldName: m.Name,
			})
		}
	}
	return stream.Send(&eventingpb.AuthEvent{
		Event: &eventingpb.AuthEvent_Presence{
			Presence: &fleetpb.PresenceEvent{
				Event: &fleetpb.PresenceEvent_ListEndpoint{
					ListEndpoint: &fleetpb.ListEndpointEvent{
						Code:      handoff.Code,
						Endpoints: ms,
					},
				},
			},
		},
	})
}
