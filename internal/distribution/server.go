package gateway

import (
	distribution "github.com/infinity-blackhole/elkia/pkg/api/distribution/v1alpha1"
)

type ServerConfig struct {
}

func NewServer(cfg ServerConfig) *Server {
	return &Server{}
}

type Server struct {
	distribution.UnimplementedClientServer
}
