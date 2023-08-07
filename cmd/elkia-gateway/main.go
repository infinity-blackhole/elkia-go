package main

import (
	"fmt"
	"net"
	"os"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/infinity-blackhole/elkia/internal/gateway"
	eventing "github.com/infinity-blackhole/elkia/pkg/api/eventing/v1alpha1"
	fleet "github.com/infinity-blackhole/elkia/pkg/api/fleet/v1alpha1"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func init() {
	if logLevelStr := os.Getenv("LOG_LEVEL"); logLevelStr != "" {
		logLevel, err := logrus.ParseLevel(logLevelStr)
		if err != nil {
			logrus.Fatal(err)
		}
		logrus.SetLevel(logLevel)
	}
}

func main() {
	fleetEndpoint := os.Getenv("ELKIA_FLEET_ENDPOINT")
	if fleetEndpoint == "" {
		fleetEndpoint = "localhost:8080"
	}
	logrus.Debugf("gateway: connecting to fleet at %s", fleetEndpoint)
	fleetConn, err := grpc.Dial(
		fleetEndpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		logrus.Fatal(err)
	}
	logrus.Debugf("gateway: connected to fleet at %s", fleetEndpoint)
	redisClient := redis.NewUniversalClient(&redis.UniversalOptions{
		Addrs:    []string{"localhost:6379"},
		Password: "",
	})
	logrus.Debugf("gateway: connected to redis")
	srv := grpc.NewServer(
		grpc.UnaryInterceptor(otelgrpc.UnaryServerInterceptor()),
		grpc.StreamInterceptor(otelgrpc.StreamServerInterceptor()),
	)
	host := os.Getenv("HOST")
	if host == "" {
		host = "localhost"
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%s", host, port))
	if err != nil {
		logrus.Fatal(err)
	}
	eventing.RegisterGatewayServer(
		srv,
		gateway.NewServer(gateway.ServerConfig{
			PresenceClient: fleet.NewPresenceClient(fleetConn),
			RedisClient:    redisClient,
		}),
	)
	logrus.Debugf("auth server: listening on %s:%s", host, port)
	if err := srv.Serve(lis); err != nil {
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
