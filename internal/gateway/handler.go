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
		fleetClient:   cfg.FleetClient,
		kafkaProducer: cfg.KafkaProducer,
		kafkaConsumer: cfg.KafkaConsumer,
	}
}

type Handler struct {
	fleetClient   fleet.FleetClient
	kafkaProducer *kafka.Producer
	kafkaConsumer *kafka.Consumer
}

func (h *Handler) ServeNosTale(c net.Conn) {
	ctx := context.Background()
	conn := h.newConn(c)
	go conn.serve(ctx)
}

func (h *Handler) newConn(c net.Conn) *Conn {
	return &Conn{
		conn:          c,
		rc:            monoalphabetic.NewReader(bufio.NewReader(c)),
		fleetClient:   h.fleetClient,
		kafkaProducer: h.kafkaProducer,
	}
}

type Conn struct {
	conn          net.Conn
	rc            *monoalphabetic.Reader
	fleetClient   fleet.FleetClient
	kafkaProducer *kafka.Producer
}

func (c *Conn) serve(ctx context.Context) {
	ack, err := c.handleHandoff(ctx)
	if err != nil {
		panic(err)
	}
	for {
		rs, err := c.newMessageReaders()
		if err != nil {
			panic(err)
		}
		for _, r := range rs {
			msg, err := r.ReadChannelMessage()
			if err != nil {
				panic(err)
			}
			if msg.Sequence != ack.Sequence+1 {
				panic(errors.New("invalid sequence number"))
			} else {
				ack.Sequence = msg.Sequence
			}
			c.kafkaProducer.Produce(&kafka.Message{
				TopicPartition: kafka.TopicPartition{
					Partition: kafka.PartitionAny,
				},
			}, nil)
		}
	}
}

func (c *Conn) handleHandoff(
	ctx context.Context,
) (*eventing.AcknowledgeHandoffMessage, error) {
	rs, err := c.newMessageReaders()
	if err != nil {
		return nil, err
	}
	syncMsg, err := rs[0].ReadSyncMessage()
	if err != nil {
		return nil, err
	}
	handoffMsg, err := rs[1].ReadPerformHandoffMessage()
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
	_, err = c.fleetClient.
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

func (c *Conn) newMessageReaders() ([]*protonostale.GatewayMessageReader, error) {
	buff, err := c.rc.ReadMessageBytes()
	if err != nil {
		return nil, err
	}
	readers := make([]*protonostale.GatewayMessageReader, len(buff))
	for i, b := range buff {
		readers[i] = protonostale.NewGatewayMessageReader(bytes.NewReader(b))
	}
	return readers, nil
}
