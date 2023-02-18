package gateway

import (
	"bufio"
	"context"
	"errors"
	"net"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	eventing "github.com/infinity-blackhole/elkia/pkg/api/eventing/v1alpha1"
	fleet "github.com/infinity-blackhole/elkia/pkg/api/fleet/v1alpha1"
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
		rwc:           c,
		rc:            protonostale.NewGatewayReader(bufio.NewReader(c)),
		wc:            protonostale.NewGatewayWriter(bufio.NewWriter(c)),
		fleetClient:   h.fleetClient,
		kafkaProducer: h.kafkaProducer,
	}
}

type Conn struct {
	rwc           net.Conn
	rc            *protonostale.GatewayReader
	wc            *protonostale.GatewayWriter
	fleetClient   fleet.FleetClient
	kafkaProducer *kafka.Producer
}

func (c *Conn) serve(ctx context.Context) {
	ack, err := c.handoff(ctx)
	if err != nil {
		panic(err)
	}
	for {
		rs, err := c.rc.ReadMessageSlice()
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

func (c *Conn) handoff(
	ctx context.Context,
) (*eventing.AcknowledgeHandoffMessage, error) {
	rs, err := c.rc.ReadMessageSlice()
	if err != nil {
		return nil, err
	}
	if len(rs) != 2 {
		return nil, errors.New("invalid handoff message")
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
	c.rc.SetKey(handoffMsg.KeyMessage.Key)
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
