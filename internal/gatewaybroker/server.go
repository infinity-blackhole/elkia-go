package gateway

import (
	"errors"
	"strconv"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	eventing "github.com/infinity-blackhole/elkia/pkg/api/eventing/v1alpha1"
	fleet "github.com/infinity-blackhole/elkia/pkg/api/fleet/v1alpha1"
	"google.golang.org/grpc/metadata"
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

type ChannelServer struct {
	eventing.GatewayBroker_ChannelInteractServer
	Sequence uint32
	Key      uint32
}

type Server struct {
	eventing.UnimplementedGatewayBrokerServer
	presence      fleet.PresenceClient
	kafkaProducer *kafka.Producer
	kafkaConsumer *kafka.Consumer
}

func (s *Server) AuthHandoffInteract(
	stream eventing.GatewayBroker_AuthHandoffInteractServer,
) error {
	m, err := stream.Recv()
	if err != nil {
		return err
	}
	sync := m.GetSyncEvent()
	if sync == nil {
		return errors.New("handoff: handshake sync protocol error")
	}
	login := m.GetLoginEvent()
	if login == nil {
		return errors.New("handoff: handshake sync protocol error")
	}
	if login.KeyEvent.Sequence != sync.Sequence+1 {
		return errors.New("handoff: handshake sync protocol error")
	}
	if login.PasswordEvent.Sequence != login.KeyEvent.Sequence+1 {
		return errors.New("handoff: handshake sync protocol error")
	}
	_, err = s.presence.AuthHandoff(stream.Context(), &fleet.AuthHandoffRequest{
		Key:      login.KeyEvent.Key,
		Password: login.PasswordEvent.Password,
	})
	if err != nil {
		return err
	}
	return stream.Send(&eventing.AuthHandoffInteractResponse{
		Payload: &eventing.AuthHandoffInteractResponse_LoginSuccessEvent{
			LoginSuccessEvent: &eventing.AuthHandoffLoginSuccessEvent{
				Key: login.KeyEvent.Key,
			},
		},
	})
}

func (s *Server) ChannelInteract(stream eventing.GatewayBroker_ChannelInteractServer) error {
	md, ok := metadata.FromIncomingContext(stream.Context())
	if !ok {
		return errors.New("channel: handshake metadata error")
	}
	sequence, err := GetSequenceFromMetadata(md)
	if err != nil {
		return err
	}
	_, err = GetKeyFromMetadata(md)
	if err != nil {
		return err
	}
	for {
		m, err := stream.Recv()
		if err != nil {
			return err
		}
		switch m.Payload {
		case &eventing.ChannelInteractRequest_ChannelEvent{}:
			m := m.GetChannelEvent()
			if sequence != m.Sequence {
				return errors.New("channel: handshake sync protocol error")
			}
			sequence++
		default:
			return errors.New("channel: handshake sync protocol error")
		}
	}
}

func GetSequenceFromMetadata(md metadata.MD) (uint32, error) {
	sequences := md.Get("sequence")
	if len(sequences) != 1 {
		return 0, errors.New("channel: handshake metadata error")
	}
	sequence, err := strconv.ParseUint(sequences[0], 10, 32)
	if err != nil {
		return 0, err
	}
	return uint32(sequence), nil
}

func GetKeyFromMetadata(md metadata.MD) (uint32, error) {
	keys := md.Get("key")
	if len(keys) != 1 {
		return 0, errors.New("channel: handshake metadata error")
	}
	key, err := strconv.ParseUint(keys[0], 10, 32)
	if err != nil {
		return 0, err
	}
	return uint32(key), nil
}
