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
	handoff := m.GetHandoffFrame()
	if handoff == nil {
		return errors.New("handoff: handshake sync protocol error")
	}
	if handoff.IdentifierFrame.Sequence != s.sequence+1 {
		return errors.New("handoff: handshake sync protocol error")
	}
	if handoff.PasswordFrame.Sequence != handoff.IdentifierFrame.Sequence+1 {
		return errors.New("handoff: handshake sync protocol error")
	}
	s.sequence = handoff.PasswordFrame.Sequence
	if handoff.IdentifierFrame.Sequence != sync.Sequence+1 {
		return errors.New("handoff: handshake sync protocol error")
	}
	if handoff.PasswordFrame.Sequence != handoff.IdentifierFrame.Sequence+1 {
		return errors.New("handoff: handshake sync protocol error")
	}
	_, err = s.presence.AuthHandoff(stream.Context(), &fleet.AuthHandoffRequest{
		Code:       sync.Code,
		Identifier: handoff.IdentifierFrame.Identifier,
		Password:   handoff.PasswordFrame.Password,
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
