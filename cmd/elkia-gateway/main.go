package main

import (
	core "github.com/infinity-blackhole/elkia/internal/app"
	app "github.com/infinity-blackhole/elkia/internal/elkia-gateway"
)

var (
	IdentityTopic = "identity"
)

func main() {
	c := new(core.Config)
	kp, err := app.NewKafkaProducer(c)
	if err != nil {
		panic(err)
	}
	defer kp.Close()
	kc, err := app.NewKafkaConsumer(c)
	if err != nil {
		panic(err)
	}
	defer kc.Close()
	kc.SubscribeTopics([]string{IdentityTopic}, nil)
	s, err := app.NewNosTaleServer(c, kp, kc)
	if err != nil {
		panic(err)
	}
	if err := s.ListenAndServe(); err != nil {
		panic(err)
	}
}
