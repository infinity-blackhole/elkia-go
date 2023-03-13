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
	msg, err := stream.Recv()
	if err != nil {
		return err
	}
	sync := msg.GetSyncFrame()
	if sync == nil {
		return errors.New("handoff: handshake sync protocol error")
	}
	s.sequence = sync.Sequence
	msg, err = stream.Recv()
	if err != nil {
		return err
	}
	identifier := msg.GetIdentifierFrame()
	if identifier.Sequence != s.sequence+1 {
		return errors.New("handoff: handshake sync protocol error")
	}
	s.sequence = identifier.Sequence
	msg, err = stream.Recv()
	if err != nil {
		return err
	}
	password := msg.GetPasswordFrame()
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
		msg, err := stream.Recv()
		if err != nil {
			return err
		}
		switch msg.Payload.(type) {
		default:
			return errors.New("channel: handshake sync protocol error")
		}
	}
}
