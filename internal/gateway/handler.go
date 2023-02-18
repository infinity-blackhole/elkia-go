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
	"github.com/infinity-blackhole/elkia/pkg/nostale/simplesubtitution"
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
	r := simplesubtitution.NewReader(bufio.NewReader(c))
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
	r *simplesubtitution.Reader,
) (*eventing.AcknowledgeHandoffMessage, error) {
	syncMessage, err := ReadSyncMessage(r)
	if err != nil {
		return nil, err
	}
	handoffMessage, err := ReadHandoffMessage(r)
	if err != nil {
		return nil, err
	}
	if syncMessage.Sequence != handoffMessage.KeySequence+1 {
		return nil, errors.New("corrupted sync message")
	}
	if handoffMessage.KeySequence != handoffMessage.PasswordSequence+1 {
		return nil, errors.New("corrupted handoff message")
	}
	if err != nil {
		panic(err)
	}
	_, err = h.fleet.
		PerformHandoff(ctx, &fleet.PerformHandoffRequest{
			Key:   handoffMessage.Key,
			Token: handoffMessage.Password,
		})
	if err != nil {
		return nil, err
	}
	return &eventing.AcknowledgeHandoffMessage{
		Key:      handoffMessage.Key,
		Sequence: syncMessage.Sequence,
	}, nil
}

func (h *Handler) readerMessage(
	r *simplesubtitution.Reader,
) (string, uint32, io.Reader, error) {
	s, err := r.ReadMessageBytes()
	if err != nil {
		panic(err)
	}
	ss := bytes.Split(s, []byte{' '})
	if len(ss) == 0 {
		panic(errors.New("invalid message"))
	}
	sequenceNumber, err := protonostale.ParseUint32(ss[0])
	if err != nil {
		panic(err)
	}
	return string(ss[0]), sequenceNumber, bytes.NewReader(ss[1]), nil
}

func ReadSyncMessage(
	r *simplesubtitution.Reader,
) (*eventing.SyncMessage, error) {
	s, err := r.ReadMessageBytes()
	if err != nil {
		return nil, err
	}
	return protonostale.ParseSyncMessage(s)
}

func ReadHandoffMessage(
	r *simplesubtitution.Reader,
) (*eventing.PerformHandoffMessage, error) {
	s, err := r.ReadMessageBytes()
	if err != nil {
		return nil, err
	}
	return protonostale.ParsePerformHandoffMessage(s)
}
