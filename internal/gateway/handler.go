package gateway

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"io"
	"net"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	eventing "github.com/infinity-blackhole/elkia/pkg/api/eventing/v1alpha1"
	fleet "github.com/infinity-blackhole/elkia/pkg/api/fleet/v1alpha1"
	"github.com/infinity-blackhole/elkia/pkg/nostale/monoalphabetic"
	"github.com/infinity-blackhole/elkia/pkg/protonostale"
)

type HandlerConfig struct {
	FleetClient   fleet.FleetClient
	KafkaProducer *kafka.Producer
	KafkaConsumer *kafka.Consumer
}

func NewHandler(cfg HandlerConfig) *Handler {
	return &Handler{
		fleet:         cfg.FleetClient,
		kafkaProducer: cfg.KafkaProducer,
		kafkaConsumer: cfg.KafkaConsumer,
	}
}

type Handler struct {
	fleet         fleet.FleetClient
	kafkaProducer *kafka.Producer
	kafkaConsumer *kafka.Consumer
}

func (h *Handler) ServeNosTale(c net.Conn) {
	ctx := context.Background()
	r := monoalphabetic.NewReader(bufio.NewReader(c))
	ack, err := h.handleHandoff(ctx, c, r)
	if err != nil {
		panic(err)
	}
	for {
		_, seqNum, _, err := h.readerMessage(r)
		if err != nil {
			panic(err)
		}
		if seqNum != ack.Sequence+1 {
			panic(errors.New("invalid sequence number"))
		} else {
			ack.Sequence = seqNum
		}
		h.kafkaProducer.Produce(&kafka.Message{
			TopicPartition: kafka.TopicPartition{
				Partition: kafka.PartitionAny,
			},
		}, nil)
	}
}

func (h *Handler) handleHandoff(
	ctx context.Context,
	c net.Conn,
	r *monoalphabetic.Reader,
) (*eventing.AcknowledgeHandoffMessage, error) {
	syncMsg, err := ReadSyncMessage(r)
	if err != nil {
		return nil, err
	}
	handoffMsg, err := ReadHandoffMessage(r)
	if err != nil {
		return nil, err
	}
	if syncMsg.Sequence != handoffMsg.KeyMessage.Sequence+1 {
		return nil, errors.New("corrupted sync message")
	}
	if handoffMsg.KeyMessage.Sequence != handoffMsg.PasswordMessage.Sequence+1 {
		return nil, errors.New("corrupted handoff message")
	}
	if err != nil {
		panic(err)
	}
	_, err = h.fleet.
		PerformHandoff(ctx, &fleet.PerformHandoffRequest{
			Key:   handoffMsg.KeyMessage.Key,
			Token: handoffMsg.PasswordMessage.Password,
		})
	if err != nil {
		return nil, err
	}
	return &eventing.AcknowledgeHandoffMessage{
		Key:      handoffMsg.KeyMessage.Key,
		Sequence: syncMsg.Sequence,
	}, nil
}

func (h *Handler) readerMessage(
	r *monoalphabetic.Reader,
) (string, uint32, io.Reader, error) {
	s, err := r.ReadMessageBytes()
	if err != nil {
		panic(err)
	}
	ss := bytes.Split(s[0], []byte{' '})
	if len(ss) == 0 {
		panic(errors.New("invalid message"))
	}
	sn, err := protonostale.ParseUint32(ss[1])
	if err != nil {
		panic(err)
	}
	return string(ss[0]), sn, bytes.NewReader(ss[2]), nil
}

func ReadSyncMessage(
	r *monoalphabetic.Reader,
) (*eventing.SyncMessage, error) {
	s, err := r.ReadMessageBytes()
	if err != nil {
		return nil, err
	}
	if len(s) != 1 {
		return nil, errors.New("invalid sync message")
	}
	return protonostale.ParseSyncMessage(s[0])
}

func ReadHandoffMessage(
	r *monoalphabetic.Reader,
) (*eventing.PerformHandoffMessage, error) {
	s, err := r.ReadMessageBytes()
	if err != nil {
		return nil, err
	}
	key, err := protonostale.ParseKeyMessage(s[0])
	if err != nil {
		return nil, err
	}
	pwd, err := protonostale.ParsePasswordMessage(s[1])
	if err != nil {
		return nil, err
	}
	return &eventing.PerformHandoffMessage{
		KeyMessage:      key,
		PasswordMessage: pwd,
	}, nil
}
