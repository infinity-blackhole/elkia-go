package main

import (
	"strings"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/infinity-blackhole/elkia/internal/gateway"
	fleetv1alpha1pb "github.com/infinity-blackhole/elkia/pkg/api/fleet/v1alpha1"
	"github.com/infinity-blackhole/elkia/pkg/nostale"
	"github.com/spf13/pflag"
	"google.golang.org/grpc"
)

var (
	address            string
	elkiaFleetEndpoint string
	kafkaEndpoints     []string
	kafkaTopics        []string
	kafkaGroupID       string
)

func init() {
	pflag.StringVar(
		&address,
		"address",
		":4124",
		"Address",
	)
	pflag.StringVar(
		&elkiaFleetEndpoint,
		"elkia-fleet-endpoint",
		"localhost:8080",
		"Elkia Fleet endpoint",
	)
	pflag.StringSliceVar(
		&kafkaEndpoints,
		"kafka-endpoints",
		[]string{"localhost:9092"},
		"Kafka endpoints",
	)
	pflag.StringSliceVar(
		&kafkaTopics,
		"kafka-topics",
		[]string{"identity"},
		"Kafka topics",
	)
	pflag.StringVar(
		&kafkaGroupID,
		"kafka-group-id",
		"elkia-gateway",
		"Kafka group ID",
	)
}

func main() {
	pflag.Parse()
	conn, err := grpc.Dial(elkiaFleetEndpoint, grpc.WithInsecure())
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
		"bootstrap.servers": strings.Join(kafkaEndpoints, ","),
	})
}

func NewKafkaConsumer() (*kafka.Consumer, error) {
	return kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": strings.Join(kafkaEndpoints, ","),
		"group.id":          kafkaGroupID,
		"auto.offset.reset": "earliest",
	})
}
