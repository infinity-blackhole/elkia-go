package gateway

import (
	"errors"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	eventing "github.com/infinity-blackhole/elkia/pkg/api/eventing/v1alpha1"
	fleet "github.com/infinity-blackhole/elkia/pkg/api/fleet/v1alpha1"
)

type ServerConfig struct {
	PresenceClient fleet.PresenceClient
	KafkaProducer  *kafka.Producer
	KafkaConsumer  *kafka.Consumer
}

func NewServer(cfg ServerConfig) *Server {
	return &Server{
		presence:      cfg.PresenceClient,
		kafkaProducer: cfg.KafkaProducer,
		kafkaConsumer: cfg.KafkaConsumer,
	}
}

type Server struct {
	eventing.UnimplementedGatewayServer
	presence      fleet.PresenceClient
	kafkaProducer *kafka.Producer
	kafkaConsumer *kafka.Consumer
	sequence      uint32
}

func (s *Server) ChannelInteract(stream eventing.Gateway_ChannelInteractServer) error {
	m, err := stream.Recv()
	if err != nil {
		return err
	}
	sync := m.GetSyncFrame()
	if sync == nil {
		return errors.New("handoff: handshake sync protocol error")
	}
	s.sequence = sync.Sequence
	m, err = stream.Recv()
	if err != nil {
		return err
	}
	identifier := m.GetIdentifierFrame()
	if identifier.Sequence != s.sequence+1 {
		return errors.New("handoff: handshake sync protocol error")
	}
	m, err = stream.Recv()
	if err != nil {
		return err
	}
	s.sequence = identifier.Sequence
	password := m.GetPasswordFrame()
	if password.Sequence != s.sequence+1 {
		return errors.New("handoff: handshake sync protocol error")
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
	for {
		m, err := stream.Recv()
		if err != nil {
			return err
		}
		switch p := m.Payload.(type) {
		case *eventing.ChannelInteractRequest_RawFrame:
			if s.sequence != p.RawFrame.Sequence {
				return errors.New("channel: handshake sync protocol error")
			}
			s.sequence++
		default:
			return errors.New("channel: handshake sync protocol error")
		}
	}
}
