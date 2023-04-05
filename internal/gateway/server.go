package gateway

import (
	"errors"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	eventing "github.com/infinity-blackhole/elkia/pkg/api/eventing/v1alpha1"
	fleet "github.com/infinity-blackhole/elkia/pkg/api/fleet/v1alpha1"
	"google.golang.org/protobuf/proto"
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
	for {
		msg, err := stream.Recv()
		if err != nil {
			return err
		}
		buff, err := proto.Marshal(msg)
		if err != nil {
			return err
		}
		if err := s.kafkaProducer.Produce(
			&kafka.Message{
				Value: buff,
				Headers: []kafka.Header{
					{
						Key:   "identity",
						Value: []byte(identifier.Identifier),
					},
				},
			},
			nil,
		); err != nil {
			return err
		}
	}
}
