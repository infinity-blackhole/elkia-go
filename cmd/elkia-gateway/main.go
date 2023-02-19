package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/infinity-blackhole/elkia/internal/gateway"
	fleet "github.com/infinity-blackhole/elkia/pkg/api/fleet/v1alpha1"
	"github.com/infinity-blackhole/elkia/pkg/nostale"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func init() {
	logrus.SetLevel(logrus.DebugLevel)
}

func main() {
	elkiaFleetEndpoint := os.Getenv("ELKIA_FLEET_ENDPOINT")
	if elkiaFleetEndpoint == "" {
		elkiaFleetEndpoint = "localhost:8080"
	}
	logrus.Debugf("gateway: connecting to fleet at %s", elkiaFleetEndpoint)
	conn, err := grpc.Dial(
		elkiaFleetEndpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		logrus.Fatal(err)
	}
	logrus.Debugf("gateway: connected to fleet at %s", elkiaFleetEndpoint)
	kp, err := NewKafkaProducer()
	if err != nil {
		logrus.Fatal(err)
	}
	logrus.Debugf("gateway: connected to kafka producer")
	defer kp.Close()
	kc, err := NewKafkaConsumer()
	if err != nil {
		logrus.Fatal(err)
	}
	logrus.Debugf("gateway: connected to kafka consumer")
	defer kc.Close()
	kafkaTopicsStr := os.Getenv("KAFKA_TOPICS")
	var kafkaTopics []string
	if kafkaTopicsStr != "" {
		kafkaTopics = strings.Split(kafkaTopicsStr, ",")
	} else {
		kafkaTopics = []string{"identity"}
	}
	logrus.Debugf("gateway: subscribing to kafka topics %v", kafkaTopics)
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
	logrus.Debugf("gateway: listening on %s:%s", host, port)
	if err := s.ListenAndServe(); err != nil {
		logrus.Fatal(err)
	}
}

func NewKafkaProducer() (*kafka.Producer, error) {
	kafkaEndpoints := os.Getenv("KAFKA_ENDPOINTS")
	if kafkaEndpoints == "" {
		kafkaEndpoints = "localhost:9092"
	}
	logrus.Debugf("gateway: connecting to kafka at %s", kafkaEndpoints)
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
	logrus.Debugf("gateway: connecting to kafka at %s", kafkaEndpoints)
	return kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": kafkaEndpoints,
		"group.id":          kafkaGroupID,
		"auto.offset.reset": "earliest",
	})
}
