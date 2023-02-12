package elkiagateway

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"net"
	"net/textproto"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/infinity-blackhole/elkia/pkg/core"
	"github.com/infinity-blackhole/elkia/pkg/crypto"
	"github.com/infinity-blackhole/elkia/pkg/messages"
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
	r := NewReader(bufio.NewReader(c))
	_, lastSequenceNumber, err := h.handleHandoff(ctx, c, r)
	if err != nil {
		panic(err)
	}
	for {
		s, err := r.ReadLine()
		if err != nil {
			panic(err)
		}
		ss := bytes.Split(s, []byte{' '})
		if len(ss) == 0 {
			panic(errors.New("invalid message"))
		}
		sequenceNumber, err := messages.ParseUint32(ss[0])
		if err != nil {
			panic(err)
		}
		if sequenceNumber != lastSequenceNumber+1 {
			panic(errors.New("invalid sequence number"))
		} else {
			lastSequenceNumber = sequenceNumber
		}
		h.kafkaProducer.Produce(&kafka.Message{
			TopicPartition: kafka.TopicPartition{
				Partition: kafka.PartitionAny,
			},
			Value: s,
		}, nil)
	}
}

func (h *Handler) handleHandoff(ctx context.Context, c net.Conn, r *Reader) (uint32, uint32, error) {
	syncMessage, err := ReadSyncMessage(r)
	if err != nil {
		return 0, 0, err
	}
	handoffMessage, err := ReadHandoffMessage(r)
	if err != nil {
		return 0, 0, err
	}
	if handoffMessage.SequenceNumber != syncMessage.SequenceNumber+1 {
		return 0, 0, errors.New("invalid sequence number")
	}
	presence, err := h.sessionStore.GetHandoffSession(ctx, handoffMessage.Key)
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
		handoffMessage.Password,
	)
	if err != nil {
		return 0, 0, err
	}
	if keySession.Identity.Id != credentialsSession.Identity.Id {
		return 0, 0, errors.New("invalid credentials")
	}
	if err := h.sessionStore.DeleteHandoffSession(ctx, handoffMessage.Key); err != nil {
		return 0, 0, err
	}
	return handoffMessage.Key, syncMessage.SequenceNumber, nil
}

type Reader struct {
	r      *textproto.Reader
	crypto *crypto.SimpleSubstitution
}

func NewReader(r *bufio.Reader) *Reader {
	return &Reader{
		r:      textproto.NewReader(r),
		crypto: new(crypto.SimpleSubstitution),
	}
}

func (r *Reader) ReadLine() ([]byte, error) {
	s, err := r.r.ReadLineBytes()
	if err != nil {
		return nil, err
	}
	return r.crypto.Decrypt(s), nil
}

func ReadSyncMessage(r *Reader) (*messages.SyncMessage, error) {
	s, err := r.ReadLine()
	if err != nil {
		return nil, err
	}
	return messages.ParseSyncMessage(s)
}

func ReadHandoffMessage(r *Reader) (*messages.HandoffMessage, error) {
	s, err := r.ReadLine()
	if err != nil {
		return nil, err
	}
	return messages.ParseHandoffMessage(s)
}
