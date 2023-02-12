package main

import (
	core "github.com/infinity-blackhole/elkia/internal/app"
	app "github.com/infinity-blackhole/elkia/internal/elkia-authserver"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

func main() {
	c := new(core.Config)
	srv, err := app.NewNosTaleServer(c)
	if err != nil {
		panic(err)
	}
	if err := srv.ListenAndServe(); err != nil {
		panic(err)
	}
}
