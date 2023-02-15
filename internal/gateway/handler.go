package gateway

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"io"
	"net"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	eventingv1alpha1pb "github.com/infinity-blackhole/elkia/pkg/api/eventing/v1alpha1"
	fleetv1alpha1pb "github.com/infinity-blackhole/elkia/pkg/api/fleet/v1alpha1"
	"github.com/infinity-blackhole/elkia/pkg/crypto"
	"github.com/infinity-blackhole/elkia/pkg/protonostale"
)

type HandlerConfig struct {
	FleetClient   fleetv1alpha1pb.FleetServiceClient
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
	fleet         fleetv1alpha1pb.FleetServiceClient
	kafkaProducer *kafka.Producer
	kafkaConsumer *kafka.Consumer
}

func (h *Handler) ServeNosTale(c net.Conn) {
	ctx := context.Background()
	r := crypto.NewServerReader(bufio.NewReader(c))
	_, lastSeqNum, err := h.handleHandoff(ctx, c, r)
	if err != nil {
		panic(err)
	}
	for {
		_, seqNum, _, err := h.readerMessage(r)
		if err != nil {
			panic(err)
		}
		if seqNum != lastSeqNum+1 {
			panic(errors.New("invalid sequence number"))
		} else {
			lastSeqNum = seqNum
		}
		h.kafkaProducer.Produce(&kafka.Message{
			TopicPartition: kafka.TopicPartition{
				Partition: kafka.PartitionAny,
			},
		}, nil)
	}
}

func (h *Handler) handleHandoff(ctx context.Context, c net.Conn, r *crypto.ServerReader) (uint32, uint32, error) {
	syncMessage, err := ReadSyncMessage(r)
	if err != nil {
		return 0, 0, err
	}
	handoffMessage, err := ReadHandoffMessage(r)
	if err != nil {
		return 0, 0, err
	}
	if syncMessage.Sequence != handoffMessage.KeySequence+1 {
		return 0, 0, errors.New("corrupted sync message")
	}
	if handoffMessage.KeySequence != handoffMessage.PasswordSequence+1 {
		return 0, 0, errors.New("corrupted handoff message")
	}
	if err != nil {
		panic(err)
	}
	_, err = h.fleet.
		PerformHandoff(ctx, &fleetv1alpha1pb.PerformHandoffRequest{
			Key:   handoffMessage.Key,
			Token: handoffMessage.Password,
		})
	if err != nil {
		return 0, 0, err
	}
	return handoffMessage.Key, syncMessage.Sequence, nil
}

func (h *Handler) readerMessage(r *crypto.ServerReader) (string, uint32, io.Reader, error) {
	s, err := r.ReadLineBytes()
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

func ReadSyncMessage(r *crypto.ServerReader) (*eventingv1alpha1pb.SyncMessage, error) {
	s, err := r.ReadLineBytes()
	if err != nil {
		return nil, err
	}
	return protonostale.ParseSyncMessage(s)
}

func ReadHandoffMessage(r *crypto.ServerReader) (*eventingv1alpha1pb.PerformHandoffMessage, error) {
	s, err := r.ReadLineBytes()
	if err != nil {
		return nil, err
	}
	return protonostale.ParsePerformHandoffMessage(s)
}
