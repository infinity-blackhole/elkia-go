package elkiagateway

import (
	"bufio"
	"context"
	"errors"
	"net"
	"net/textproto"
	"strings"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/infinity-blackhole/elkia/pkg/core"
)

type HandlerConfig struct {
	IAMClient     *core.IdentityProvider
	SessionStore  *core.SessionStore
	Fleet         *core.FleetClient
	KafkaProducer *kafka.Producer
	KafkaConsumer *kafka.Consumer
}

func NewHandler(cfg HandlerConfig) *Handler {
	return &Handler{
		identityProvider: cfg.IAMClient,
		sessionStore:     cfg.SessionStore,
		fleet:            cfg.Fleet,
		kafkaProducer:    cfg.KafkaProducer,
		kafkaConsumer:    cfg.KafkaConsumer,
	}
}

type Handler struct {
	identityProvider *core.IdentityProvider
	sessionStore     *core.SessionStore
	fleet            *core.FleetClient
	kafkaProducer    *kafka.Producer
	kafkaConsumer    *kafka.Consumer
}

func (h *Handler) ServeNosTale(c net.Conn) {
	ctx := context.Background()
	r := textproto.NewReader(bufio.NewReader(c))
	_, lastSN, err := h.handleHandoff(ctx, c, r)
	if err != nil {
		panic(err)
	}
	for {
		s, err := r.ReadLine()
		if err != nil {
			panic(err)
		}
		ss := strings.Split(s, " ")
		if 0 >= len(ss) {
			panic(errors.New("invalid message"))
		}
		sn, err := ParseUint32(ss[0])
		if err != nil {
			panic(err)
		}
		if sn != lastSN+1 {
			panic(errors.New("invalid sequence number"))
		} else {
			lastSN = sn
		}
		h.kafkaProducer.Produce(&kafka.Message{
			TopicPartition: kafka.TopicPartition{
				Partition: kafka.PartitionAny,
			},
			Value: []byte(s),
		}, nil)
	}
}

func (h *Handler) handleHandoff(ctx context.Context, c net.Conn, r *textproto.Reader) (uint32, uint32, error) {
	s, err := r.ReadLine()
	if err != nil {
		return 0, 0, err
	}
	syncMessage, err := ParseSyncMessage(s)
	if err != nil {
		return 0, 0, err
	}
	s, err = r.ReadLine()
	if err != nil {
		return 0, 0, err
	}
	credsMessage, err := ParseCredentialsMessage(s)
	if err != nil {
		return 0, 0, err
	}
	if credsMessage.SequenceNumber != syncMessage.SequenceNumber+1 {
		return 0, 0, errors.New("invalid sequence number")
	}
	presence, err := h.sessionStore.GetHandoffSession(ctx, credsMessage.Key)
	if err != nil {
		return 0, 0, err
	}
	keySession, err := h.identityProvider.GetSession(ctx, presence.SessionToken)
	if err != nil {
		return 0, 0, err
	}
	if err := h.identityProvider.Logout(ctx, presence.SessionToken); err != nil {
		return 0, 0, err
	}
	credentialsSession, _, err := h.identityProvider.PerformLoginFlowWithPasswordMethod(
		ctx,
		presence.Identifier,
		credsMessage.Password,
	)
	if err != nil {
		return 0, 0, err
	}
	if keySession.Identity.Id != credentialsSession.Identity.Id {
		return 0, 0, errors.New("invalid credentials")
	}
	if err := h.sessionStore.DeleteHandoffSession(ctx, credsMessage.Key); err != nil {
		return 0, 0, err
	}
	return credsMessage.Key, syncMessage.SequenceNumber, nil
}
