package gateway

import (
	"bufio"
	"bytes"
	"context"
	"errors"
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
	r := NewMessageReader(monoalphabetic.NewReader(bufio.NewReader(c)))
	ack, err := h.handleHandoff(ctx, c, r)
	if err != nil {
		panic(err)
	}
	for {
		msgs, err := r.ReaderMessage()
		if err != nil {
			panic(err)
		}
		for _, msg := range msgs {
			if msg.Sequence != ack.Sequence+1 {
				panic(errors.New("invalid sequence number"))
			} else {
				ack.Sequence = msg.Sequence
			}
			h.kafkaProducer.Produce(&kafka.Message{
				TopicPartition: kafka.TopicPartition{
					Partition: kafka.PartitionAny,
				},
			}, nil)
		}
	}
}

func (h *Handler) handleHandoff(
	ctx context.Context,
	c net.Conn,
	r *MessageReader,
) (*eventing.AcknowledgeHandoffMessage, error) {
	syncMsg, err := r.ReadSyncMessage()
	if err != nil {
		return nil, err
	}
	handoffMsg, err := r.ReadHandoffMessage()
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

type MessageReader struct {
	R *monoalphabetic.Reader
}

func NewMessageReader(r *monoalphabetic.Reader) *MessageReader {
	return &MessageReader{
		R: r,
	}
}

func (r *MessageReader) ReadSyncMessage() (*eventing.SyncMessage, error) {
	s, err := r.R.ReadMessageBytes()
	if err != nil {
		return nil, err
	}
	if len(s) != 1 {
		return nil, errors.New("invalid sync message")
	}
	return protonostale.ParseSyncMessage(s[0])
}

func (r *MessageReader) ReadHandoffMessage() (*eventing.PerformHandoffMessage, error) {
	s, err := r.R.ReadMessageBytes()
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

func (r *MessageReader) ReaderMessage() ([]*eventing.GenericMessage, error) {
	ss, err := r.R.ReadMessageBytes()
	if err != nil {
		return nil, err
	}
	var msgs []*eventing.GenericMessage
	for _, s := range ss {
		s := bytes.Split(s, []byte{' '})
		sn, err := protonostale.ParseUint32(ss[0])
		if err != nil {
			return nil, err
		}
		msgs = append(msgs, &eventing.GenericMessage{
			Sequence: sn,
			Opcode:   string(s[1]),
			Payload:  s[:2],
		})
	}
	return msgs, nil
}
