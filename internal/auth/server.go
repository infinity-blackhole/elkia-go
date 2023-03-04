package auth

import (
	"math"
	"net"

	eventing "github.com/infinity-blackhole/elkia/pkg/api/eventing/v1alpha1"
	fleet "github.com/infinity-blackhole/elkia/pkg/api/fleet/v1alpha1"
	"github.com/sirupsen/logrus"
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
		case *eventing.AuthInteractRequest_LoginFrame:
			logrus.Debugf("auth: handle handoff")
			if err := s.AuthLoginFrameProduce(
				m.GetLoginFrame(),
				stream,
			); err != nil {
				return err
			}
		default:
			return status.Newf(codes.Unimplemented, "unimplemented frame type: %T", m.Payload).Err()
		}
	}
}

func (s *Server) AuthLoginFrameProduce(
	m *eventing.AuthLoginFrame,
	stream eventing.Auth_AuthLoginFrameProduceServer,
) error {
	handoff, err := s.presence.AuthLogin(
		stream.Context(),
		&fleet.AuthLoginRequest{
			Identifier: m.Identifier,
			Password:   m.Password,
		},
	)
	if err != nil {
		return err
	}
	logrus.Debugf("auth: create handoff: %v", handoff)

	MemberList, err := s.cluster.MemberList(
		stream.Context(),
		&fleet.MemberListRequest{},
	)
	if err != nil {
		return err
	}
	logrus.Debugf("auth: list members: %v", MemberList)
	ms := []*eventing.Endpoint{}
	for _, m := range MemberList.Members {
		host, port, err := net.SplitHostPort(m.Address)
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
	return stream.Send(&eventing.AuthInteractResponse{
		Payload: &eventing.AuthInteractResponse_EndpointListFrame{
			EndpointListFrame: &eventing.EndpointListFrame{
				Code:      handoff.Code,
				Endpoints: ms,
			},
		},
	})
}
