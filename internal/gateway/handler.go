package gateway

import (
	"bufio"
	"context"
	"net"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	eventing "github.com/infinity-blackhole/elkia/pkg/api/eventing/v1alpha1"
	fleet "github.com/infinity-blackhole/elkia/pkg/api/fleet/v1alpha1"
	"github.com/infinity-blackhole/elkia/pkg/protonostale"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
)

var name = "github.com/infinity-blackhole/elkia/internal/gateway"

type HandlerConfig struct {
	PresenceClient fleet.PresenceClient
	KafkaProducer  *kafka.Producer
	KafkaConsumer  *kafka.Consumer
}

func NewHandler(cfg HandlerConfig) *Handler {
	return &Handler{
		presence:      cfg.PresenceClient,
		kafkaProducer: cfg.KafkaProducer,
		kafkaConsumer: cfg.KafkaConsumer,
	}
}

type Handler struct {
	presence      fleet.PresenceClient
	kafkaProducer *kafka.Producer
	kafkaConsumer *kafka.Consumer
}

func (h *Handler) ServeNosTale(c net.Conn) {
	ctx := context.Background()
	conn := h.newConn(c)
	logrus.Debugf("gateway: new connection from %s", c.RemoteAddr().String())
	go conn.serve(ctx)
}

func (h *Handler) newConn(c net.Conn) *Conn {
	return &Conn{
		rwc:           c,
		rc:            protonostale.NewGatewayReader(bufio.NewReader(c)),
		wc:            protonostale.NewGatewayWriter(bufio.NewWriter(c)),
		presence:      h.presence,
		kafkaProducer: h.kafkaProducer,
	}
}

type Conn struct {
	rwc           net.Conn
	rc            *protonostale.GatewayReader
	wc            *protonostale.GatewayWriter
	presence      fleet.PresenceClient
	kafkaProducer *kafka.Producer
}

func (c *Conn) serve(ctx context.Context) {
	ctx, span := otel.Tracer(name).Start(ctx, "Serve")
	defer span.End()
	ack, err := c.handoff(ctx)
	if err != nil {
		if err := c.wc.WriteFailCodeMessage(&eventing.FailureMessage{
			Code: eventing.FailureCode_UNEXPECTED_ERROR,
		}); err != nil {
			logrus.Fatal(err)
		}
		logrus.Debug(err)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return
	}
	if ack == nil {
		if err := c.wc.WriteFailCodeMessage(&eventing.FailureMessage{
			Code: eventing.FailureCode_CANT_AUTHENTICATE,
		}); err != nil {
			logrus.Fatal(err)
		}
		logrus.Debugf("gateway: handoff failed: %v", ack)
		return
	}
	logrus.Debugf("gateway: handoff acknowledged: %v", ack)
	for {
		rs, err := c.rc.ReadMessageSlice()
		if err != nil {
			if err := c.wc.WriteFailCodeMessage(&eventing.FailureMessage{
				Code: eventing.FailureCode_UNEXPECTED_ERROR,
			}); err != nil {
				logrus.Fatal(err)
			}
			logrus.Debug(err)
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return
		}
		for _, r := range rs {
			opcode, err := r.ReadOpcode()
			if err != nil {
				if err := c.wc.WriteFailCodeMessage(&eventing.FailureMessage{
					Code: eventing.FailureCode_BAD_CASE,
				}); err != nil {
					logrus.Fatal(err)
				}
				logrus.Debug(err)
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
				return
			}
			logrus.Debugf("gateway: received opcode: %v", opcode)
			switch opcode {
			default:
				msg, err := r.ReadChannelMessage()
				if err != nil {
					if err := c.wc.WriteFailCodeMessage(&eventing.FailureMessage{
						Code: eventing.FailureCode_UNEXPECTED_ERROR,
					}); err != nil {
						logrus.Fatal(err)
					}
					logrus.Debug(err)
					span.RecordError(err)
					span.SetStatus(codes.Error, err.Error())
					return
				}
				logrus.Debugf("gateway: received message: %v", msg)
				if msg.Sequence != ack.Sequence+1 {
					if err := c.wc.WriteFailCodeMessage(&eventing.FailureMessage{
						Code: eventing.FailureCode_BAD_CASE,
					}); err != nil {
						logrus.Fatal(err)
					}
					logrus.Debugf("gateway: sequence mismatch: %v", msg)
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
	handoffMsg, err := rs[1].ReadAuthHandoffMessage()
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
	_, err = c.presence.AuthHandoff(ctx, &fleet.AuthHandoffRequest{
		Key:      handoffMsg.KeyMessage.Key,
		Password: handoffMsg.PasswordMessage.Password,
	})
	if err != nil {
		return nil, nil
	}
	return &eventing.AcknowledgeHandoffMessage{
		Key:      handoffMsg.KeyMessage.Key,
		Sequence: syncMsg.Sequence,
	}, nil
}
