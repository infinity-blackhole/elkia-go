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
	handshaker := h.newHandshaker(c)
	logrus.Debugf("gateway: new handshaker from %v", c.RemoteAddr())
	go handshaker.handshake(ctx)
}

func (h *Handler) newHandshaker(c net.Conn) *handshaker {
	return &handshaker{
		rwc:           c,
		rc:            protonostale.NewGatewayHandshakeReader(bufio.NewReader(c)),
		wc:            protonostale.NewGatewayWriter(bufio.NewWriter(c)),
		presence:      h.presence,
		kafkaProducer: h.kafkaProducer,
	}
}

type handshaker struct {
	rwc           net.Conn
	rc            *protonostale.GatewayHandshakeReader
	wc            *protonostale.GatewayWriter
	presence      fleet.PresenceClient
	kafkaProducer *kafka.Producer
}

func (c *handshaker) handshake(ctx context.Context) {
	go func() {
		if err := recover(); err != nil {
			c.wc.WriteDialogErrorEvent(&eventing.DialogErrorEvent{
				Code: eventing.DialogErrorCode_UNEXPECTED_ERROR,
			})
		}
	}()
	ctx, span := otel.Tracer(name).Start(ctx, "Serve")
	defer span.End()
	ack, err := c.handoff(ctx)
	if err != nil {
		logrus.Debugf("gateway: handoff failed: %v", err)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return
	}
	if ack == nil {
		if err := c.wc.WriteDialogErrorEvent(&eventing.DialogErrorEvent{
			Code: eventing.DialogErrorCode_CANT_AUTHENTICATE,
		}); err != nil {
			logrus.Fatal(err)
		}
		logrus.Debugf("gateway: handoff failed: %v", ack)
		return
	}
	logrus.Debugf("gateway: handoff acknowledged: %v", ack)
	conn := c.newConn(ack)
	conn.serve(ctx)
}

func (c *handshaker) handoff(
	ctx context.Context,
) (*eventing.AuthHandoffSuccessEvent, error) {
	rs, err := c.rc.ReadMessageSlice()
	if err != nil {
		return nil, err
	}
	if len(rs) != 2 {
		if err := c.wc.WriteDialogErrorEvent(&eventing.DialogErrorEvent{
			Code: eventing.DialogErrorCode_BAD_CASE,
		}); err != nil {
			return nil, err
		}
		return nil, nil
	}
	syncMsg, err := rs[0].ReadSyncEvent()
	if err != nil {
		if err := c.wc.WriteDialogErrorEvent(&eventing.DialogErrorEvent{
			Code: eventing.DialogErrorCode_BAD_CASE,
		}); err != nil {
			return nil, err
		}
		return nil, nil
	}
	handoffMsg, err := rs[1].ReadAuthHandoffEvent()
	if err != nil {
		if err := c.wc.WriteDialogErrorEvent(&eventing.DialogErrorEvent{
			Code: eventing.DialogErrorCode_BAD_CASE,
		}); err != nil {
			return nil, err
		}
		return nil, nil
	}
	if syncMsg.Sequence != handoffMsg.KeyEvent.Sequence+1 {
		if err := c.wc.WriteDialogErrorEvent(&eventing.DialogErrorEvent{
			Code: eventing.DialogErrorCode_BAD_CASE,
		}); err != nil {
			return nil, err
		}
		return nil, nil
	}
	if handoffMsg.KeyEvent.Sequence != handoffMsg.PasswordEvent.Sequence+1 {
		if err := c.wc.WriteDialogErrorEvent(&eventing.DialogErrorEvent{
			Code: eventing.DialogErrorCode_BAD_CASE,
		}); err != nil {
			return nil, err
		}
		return nil, nil
	}
	_, err = c.presence.AuthHandoff(ctx, &fleet.AuthHandoffRequest{
		Key:      handoffMsg.KeyEvent.Key,
		Password: handoffMsg.PasswordEvent.Password,
	})
	if err != nil {
		return nil, nil
	}
	return &eventing.AuthHandoffSuccessEvent{
		Key:      handoffMsg.KeyEvent.Key,
		Sequence: syncMsg.Sequence,
	}, nil
}

func (h *handshaker) newConn(ack *eventing.AuthHandoffSuccessEvent) *conn {
	return &conn{
		rwc:           h.rwc,
		rc:            protonostale.NewGatewayChannelReader(bufio.NewReader(h.rwc), ack.Key),
		wc:            protonostale.NewGatewayWriter(bufio.NewWriter(h.rwc)),
		presence:      h.presence,
		kafkaProducer: h.kafkaProducer,
		lastSequence:  ack.Sequence,
	}
}

type conn struct {
	rwc           net.Conn
	rc            *protonostale.GatewayChannelReader
	wc            *protonostale.GatewayWriter
	presence      fleet.PresenceClient
	kafkaProducer *kafka.Producer
	lastSequence  uint32
}

func (c *conn) serve(ctx context.Context) {
	_, span := otel.Tracer(name).Start(ctx, "Serve")
	defer span.End()
	for {
		rs, err := c.rc.ReadMessageSlice()
		if err != nil {
			logrus.Debugf("gateway: read failed: %v", err)
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return
		}
		for _, r := range rs {
			msg, err := r.ReadChannelEvent()
			if err != nil {
				logrus.Debugf("gateway: read failed: %v", err)
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
				return
			}
			logrus.Debugf("gateway: received message: %v", msg)
			if msg.Sequence != c.lastSequence+1 {
				if err := c.wc.WriteDialogErrorEvent(&eventing.DialogErrorEvent{
					Code: eventing.DialogErrorCode_BAD_CASE,
				}); err != nil {
					logrus.Fatal(err)
				}
				logrus.Debugf("gateway: sequence mismatch: %v", msg)
				return
			} else {
				c.lastSequence = msg.Sequence
			}
			c.kafkaProducer.Produce(&kafka.Message{
				TopicPartition: kafka.TopicPartition{
					Partition: kafka.PartitionAny,
				},
			}, nil)
		}
	}
}
