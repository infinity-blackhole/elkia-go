package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/infinity-blackhole/elkia/internal/gateway"
	fleet "github.com/infinity-blackhole/elkia/pkg/api/fleet/v1alpha1"
	"github.com/infinity-blackhole/elkia/pkg/nostale"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	elkiaFleetEndpoint := os.Getenv("ELKIA_FLEET_ENDPOINT")
	if elkiaFleetEndpoint == "" {
		elkiaFleetEndpoint = "localhost:8080"
	}
	conn, err := grpc.Dial(
		elkiaFleetEndpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
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
	kafkaTopicsStr := os.Getenv("KAFKA_TOPICS")
	var kafkaTopics []string
	if kafkaTopicsStr != "" {
		kafkaTopics = strings.Split(kafkaTopicsStr, ",")
	} else {
		kafkaTopics = []string{"identity"}
	}
	kc.SubscribeTopics(kafkaTopics, nil)
	host := os.Getenv("HOST")
	if host == "" {
		host = "localhost"
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "4123"
	}
	s := nostale.NewServer(nostale.ServerConfig{
		Addr: fmt.Sprintf("%s:%s", host, port),
		Handler: gateway.NewHandler(gateway.HandlerConfig{
			FleetClient:   fleet.NewFleetClient(conn),
			KafkaProducer: kp,
			KafkaConsumer: kc,
		}),
	})
	if err := s.ListenAndServe(); err != nil {
		panic(err)
	}
}

func NewKafkaProducer() (*kafka.Producer, error) {
	kafkaEndpoints := os.Getenv("KAFKA_ENDPOINTS")
	if kafkaEndpoints == "" {
		kafkaEndpoints = "localhost:9092"
	}
	return kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": kafkaEndpoints,
	})
}

func NewKafkaConsumer() (*kafka.Consumer, error) {
	kafkaEndpoints := os.Getenv("KAFKA_ENDPOINTS")
	if kafkaEndpoints == "" {
		kafkaEndpoints = "localhost:9092"
	}
	kafkaGroupID := os.Getenv("KAFKA_GROUP_ID")
	if kafkaGroupID == "" {
		kafkaGroupID = "elkia-gateway"
	}
	return kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": kafkaEndpoints,
		"group.id":          kafkaGroupID,
		"auto.offset.reset": "earliest",
	})
}
