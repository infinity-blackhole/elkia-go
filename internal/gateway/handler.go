package gateway

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"io"
	"net"
	"net/textproto"

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
	r := NewReader(bufio.NewReader(c))
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

func (h *Handler) handleHandoff(ctx context.Context, c net.Conn, r *Reader) (uint32, uint32, error) {
	syncMessage, err := ReadSyncMessage(r)
	if err != nil {
		return 0, 0, err
	}
	handoffMessage, err := ReadHandoffMessage(r)
	if err != nil {
		return 0, 0, err
	}
	if handoffMessage.Sequence != syncMessage.Sequence+1 {
		return 0, 0, errors.New("invalid sequence number")
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

func (h *Handler) readerMessage(r *Reader) (string, uint32, io.Reader, error) {
	s, err := r.ReadLine()
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

func ReadSyncMessage(r *Reader) (*eventingv1alpha1pb.SyncMessage, error) {
	s, err := r.ReadLine()
	if err != nil {
		return nil, err
	}
	return protonostale.ParseSyncMessage(s)
}

func ReadHandoffMessage(r *Reader) (*eventingv1alpha1pb.PerformHandoffMessage, error) {
	s, err := r.ReadLine()
	if err != nil {
		return nil, err
	}
	return protonostale.ParsePerformHandoffMessage(s)
}
