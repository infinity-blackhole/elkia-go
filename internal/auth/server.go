package auth

import (
	"math"
	"net"

	"github.com/sirupsen/logrus"
	eventing "go.shikanime.studio/elkia/pkg/api/eventing/v1alpha1"
	fleet "go.shikanime.studio/elkia/pkg/api/fleet/v1alpha1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ServerConfig struct {
	PresenceClient fleet.PresenceClient
	ClusterClient  fleet.ClusterClient
}

func NewServer(cfg ServerConfig) *Server {
	return &Server{
		presence: cfg.PresenceClient,
		cluster:  cfg.ClusterClient,
	}
}

type Server struct {
	eventing.UnimplementedAuthServer
	presence fleet.PresenceClient
	cluster  fleet.ClusterClient
}

func (s *Server) AuthInteract(stream eventing.Auth_AuthInteractServer) error {
	for {
		m, err := stream.Recv()
		if err != nil {
			return err
		}
		switch m.Payload.(type) {
		case *eventing.AuthInteractRequest_LoginCommand:
			logrus.Debugf("auth: handle handoff")
			if err := s.ProduceCreateHandoffFlowCommand(
				m.GetLoginCommand(),
				stream,
			); err != nil {
				return err
			}
		default:
			return status.Newf(codes.Unimplemented, "unimplemented frame type: %T", m.Payload).Err()
		}
	}
}

func (s *Server) ProduceCreateHandoffFlowCommand(
	m *eventing.CreateHandoffFlowCommand,
	stream eventing.Auth_AuthCreateHandoffFlowCommandProduceServer,
) error {
	handoff, err := s.presence.AuthCreateHandoffFlow(
		stream.Context(),
		&fleet.AuthCreateHandoffFlowRequest{
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
		&fleet.MemberListRequest{},
	)
	if err != nil {
		return err
	}
	logrus.Debugf("auth: list members: %v", memberList)
	var ms []*eventing.Endpoint
	for _, m := range memberList.Members {
		for _, a := range m.Addresses {
			host, port, err := net.SplitHostPort(a)
			if err != nil {
				return err
			}
			ms = append(ms, &eventing.Endpoint{
				Host:      host,
				Port:      port,
				Weight:    uint32(math.Round(float64(m.Population)/float64(m.Capacity)*20) + 1),
				WorldId:   m.WorldId,
				ChannelId: m.ChannelId,
				WorldName: m.Name,
			})
		}
	}
	return stream.Send(&eventing.AuthInteractResponse{
		Payload: &eventing.AuthInteractResponse_EndpointListEvent{
			EndpointListEvent: &eventing.EndpointListEvent{
				Code:      handoff.Code,
				Endpoints: ms,
			},
		},
	})
}
