package main

import (
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/infinity-blackhole/elkia/internal/gateway"
	fleetv1alpha1pb "github.com/infinity-blackhole/elkia/pkg/api/fleet/v1alpha1"
	"github.com/infinity-blackhole/elkia/pkg/nostale"
	"github.com/spf13/pflag"
	"google.golang.org/grpc"
)

var (
	address               string
	elkiaApiServerAddress string
	kafkaTopics           []string
)

func init() {
	pflag.StringVar(
		&address,
		"address",
		":4124",
		"Address",
	)
	pflag.StringVar(
		&elkiaApiServerAddress,
		"elkia-api-server-address",
		"localhost:8080",
		"Elkia API Server address",
	)
	pflag.StringSliceVar(
		&kafkaTopics,
		"kafka-topics",
		[]string{"identity"},
		"Kafka topics",
	)
}

func main() {
	conn, err := grpc.Dial(elkiaApiServerAddress, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	kp, err := NewKafkaProducer()
	if err != nil {
		panic(err)
	}
	defer kp.Close()
	kc, err := NewKafkaConsumer()
	if err != nil {
		panic(err)
	}
	defer kc.Close()
	kc.SubscribeTopics(kafkaTopics, nil)
	s := nostale.NewServer(nostale.ServerConfig{
		Addr: address,
		Handler: gateway.NewHandler(gateway.HandlerConfig{
			FleetClient:   fleetv1alpha1pb.NewFleetServiceClient(conn),
			KafkaProducer: kp,
			KafkaConsumer: kc,
		}),
	})
	if err := s.ListenAndServe(); err != nil {
		panic(err)
	}
}

func NewKafkaProducer() (*kafka.Producer, error) {
	return kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": "localhost",
	})
}

func NewKafkaConsumer() (*kafka.Consumer, error) {
	return kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": "localhost",
		"group.id":          "myGroup",
		"auto.offset.reset": "earliest",
	})
}
