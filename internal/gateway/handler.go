package gateway

import (
	"bufio"
	"context"
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
		if err := c.wc.WriteFailCodeMessage(&eventing.FailureMessage{
			Code: eventing.FailureCode_UNEXPECTED_ERROR,
		}); err != nil {
			panic(err)
		}
		return
	}
	if ack == nil {
		if err := c.wc.WriteFailCodeMessage(&eventing.FailureMessage{
			Code: eventing.FailureCode_CANT_AUTHENTICATE,
		}); err != nil {
			panic(err)
		}
		return
	}
	for {
		rs, err := c.rc.ReadMessageSlice()
		if err != nil {
			if err := c.wc.WriteFailCodeMessage(&eventing.FailureMessage{
				Code: eventing.FailureCode_UNEXPECTED_ERROR,
			}); err != nil {
				panic(err)
			}
			return
		}
		for _, r := range rs {
			msg, err := r.ReadChannelMessage()
			if err != nil {
				if err := c.wc.WriteFailCodeMessage(&eventing.FailureMessage{
					Code: eventing.FailureCode_UNEXPECTED_ERROR,
				}); err != nil {
					panic(err)
				}
				return
			}
			if msg.Sequence != ack.Sequence+1 {
				if err := c.wc.WriteFailCodeMessage(&eventing.FailureMessage{
					Code: eventing.FailureCode_BAD_CASE,
				}); err != nil {
					panic(err)
				}
				return
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
		if err := c.wc.WriteFailCodeMessage(&eventing.FailureMessage{
			Code: eventing.FailureCode_BAD_CASE,
		}); err != nil {
			return nil, err
		}
		return nil, nil
	}
	syncMsg, err := rs[0].ReadSyncMessage()
	if err != nil {
		if err := c.wc.WriteFailCodeMessage(&eventing.FailureMessage{
			Code: eventing.FailureCode_BAD_CASE,
		}); err != nil {
			return nil, err
		}
		return nil, nil
	}
	handoffMsg, err := rs[1].ReadPerformHandoffMessage()
	if err != nil {
		if err := c.wc.WriteFailCodeMessage(&eventing.FailureMessage{
			Code: eventing.FailureCode_BAD_CASE,
		}); err != nil {
			return nil, err
		}
		return nil, nil
	}
	if syncMsg.Sequence != handoffMsg.KeyMessage.Sequence+1 {
		if err := c.wc.WriteFailCodeMessage(&eventing.FailureMessage{
			Code: eventing.FailureCode_BAD_CASE,
		}); err != nil {
			return nil, err
		}
		return nil, nil
	}
	if handoffMsg.KeyMessage.Sequence != handoffMsg.PasswordMessage.Sequence+1 {
		if err := c.wc.WriteFailCodeMessage(&eventing.FailureMessage{
			Code: eventing.FailureCode_BAD_CASE,
		}); err != nil {
			return nil, err
		}
		return nil, nil
	}
	c.rc.SetKey(handoffMsg.KeyMessage.Key)
	_, err = c.fleetClient.
		PerformHandoff(ctx, &fleet.PerformHandoffRequest{
			Key:   handoffMsg.KeyMessage.Key,
			Token: handoffMsg.PasswordMessage.Password,
		})
	if err != nil {
		return nil, nil
	}
	return &eventing.AcknowledgeHandoffMessage{
		Key:      handoffMsg.KeyMessage.Key,
		Sequence: syncMsg.Sequence,
	}, nil
}
