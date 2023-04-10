package gateway

import (
	"errors"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	eventing "github.com/infinity-blackhole/elkia/pkg/api/eventing/v1alpha1"
	fleet "github.com/infinity-blackhole/elkia/pkg/api/fleet/v1alpha1"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/proto"
)

var (
	ChannelTopic = "channel"
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
	logrus.Debugf("gateway: received identifier frame")
	identifier := msg.GetIdentifierFrame()
	if identifier.Sequence != s.sequence+1 {
		return errors.New("handoff: session protocol error")
	}
	s.sequence = identifier.Sequence
	msg, err = stream.Recv()
	if err != nil {
		return err
	}
	logrus.Debugf("gateway: received password frame")
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
	logrus.Debugf("gateway: auth handoff successful")
	wg := &errgroup.Group{}
	wg.Go(func() error {
		return s.ChannelWatch(
			&eventing.ChannelWatchRequest{
				Sequence:   sync.Sequence,
				Code:       sync.Code,
				Identifier: identifier.Identifier,
				Password:   password.Password,
			},
			stream,
		)
	})
	wg.Go(func() error {
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
					TopicPartition: kafka.TopicPartition{
						Topic:     &ChannelTopic,
						Partition: kafka.PartitionAny,
					},
					Value: buff,
				},
				nil,
			); err != nil {
				return err
			}
			logrus.Debugf("gateway: sent message to kafka")
		}
	})
	return wg.Wait()
}

func (s *Server) ChannelWatch(
	in *eventing.ChannelWatchRequest,
	stream eventing.Gateway_ChannelWatchServer,
) error {
	for {
		msg, err := s.kafkaConsumer.ReadMessage(-1)
		if err != nil {
			return err
		}
		var res eventing.ChannelInteractResponse
		if err := proto.Unmarshal(msg.Value, &res); err != nil {
			return err
		}
		if err := stream.Send(&res); err != nil {
			return err
		}
	}
}
