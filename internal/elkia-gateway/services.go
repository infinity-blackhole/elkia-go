package elkiagateway

import (
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	core "github.com/infinity-blackhole/elkia/internal/app"
	"github.com/infinity-blackhole/elkia/pkg/nostale"
)

func NewKafkaProducer(c *core.Config) (*kafka.Producer, error) {
	return kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": "localhost",
	})
}

func NewKafkaConsumer(c *core.Config) (*kafka.Consumer, error) {
	return kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": "localhost",
		"group.id":          "myGroup",
		"auto.offset.reset": "earliest",
	})
}

func NewNosTaleServer(
	c *core.Config,
	kp *kafka.Producer,
	kc *kafka.Consumer,
) (*nostale.Server, error) {
	sm, err := core.NewSessionStoreClient(c)
	if err != nil {
		return nil, err
	}
	f, err := core.NewFleetClient(c)
	if err != nil {
		return nil, err
	}
	handler := NewHandler(HandlerConfig{
		IAMClient:     core.NewIdentityProviderClient(c),
		SessionStore:  sm,
		Fleet:         f,
		KafkaProducer: kp,
		KafkaConsumer: kc,
	})
	if err != nil {
		return nil, err
	}
	return nostale.NewServer(nostale.ServerConfig{
		Addr:    ":8080",
		Handler: handler,
	}), nil
}
