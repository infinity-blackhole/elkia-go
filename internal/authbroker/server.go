package authbroker

import (
	"math"
	"net"

	eventing "github.com/infinity-blackhole/elkia/pkg/api/eventing/v1alpha1"
	fleet "github.com/infinity-blackhole/elkia/pkg/api/fleet/v1alpha1"
	"github.com/sirupsen/logrus"
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
	eventing.UnimplementedAuthBrokerServer
	presence fleet.PresenceClient
	cluster  fleet.ClusterClient
}

func (s *Server) AuthInteract(stream eventing.AuthBroker_AuthInteractServer) error {
	for {
		event, err := stream.Recv()
		if err != nil {
			return err
		}
		switch event.Payload {
		case &eventing.AuthInteractRequest_AuthLoginEvent{}:
			logrus.Debugf("auth: handle handoff")
			if err := s.AuthLoginProduce(
				event.GetAuthLoginEvent(),
				stream,
			); err != nil {
				return err
			}
		default:
			logrus.Debugf("auth: unimplemented event: %v", event)
		}
	}
}

func (s *Server) AuthLoginProduce(
	m *eventing.AuthLoginEvent,
	stream eventing.AuthBroker_AuthLoginProduceServer,
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
	gateways := []*eventing.Gateway{}
	for _, m := range MemberList.Members {
		host, port, err := net.SplitHostPort(m.Address)
		if err != nil {
			return err
		}
		gateways = append(gateways, &eventing.Gateway{
			Host:      host,
			Port:      port,
			Weight:    uint32(math.Round(float64(m.Population)/float64(m.Capacity)*20) + 1),
			WorldId:   m.WorldId,
			ChannelId: m.ChannelId,
			WorldName: m.Name,
		})
	}
	return stream.Send(&eventing.AuthInteractResponse{
		Payload: &eventing.AuthInteractResponse_GatewayListEvent{
			GatewayListEvent: &eventing.GatewayListEvent{
				Key:      handoff.Key,
				Gateways: gateways,
			},
		},
	})
}
