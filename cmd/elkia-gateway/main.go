package main

import (
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/infinity-blackhole/elkia/internal/gateway"
	fleetv1alpha1pb "github.com/infinity-blackhole/elkia/pkg/api/fleet/v1alpha1"
	"github.com/infinity-blackhole/elkia/pkg/nostale"
	"google.golang.org/grpc"
)

var (
	IdentityTopic = "identity"
)

func main() {
	conn, err := grpc.Dial("unix:///var/run/elkia.sock", grpc.WithInsecure())
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
	kc.SubscribeTopics([]string{IdentityTopic}, nil)
	s := nostale.NewServer(nostale.ServerConfig{
		Addr: ":8080",
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
