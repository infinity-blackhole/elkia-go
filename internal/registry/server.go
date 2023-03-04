package registry

import (
	"context"

	distribution "github.com/infinity-blackhole/elkia/pkg/api/distribution/v1alpha1"
)

type ServerConfig struct {
}

func NewServer(cfg ServerConfig) *Server {
	return &Server{}
}

type Server struct {
	distribution.UnimplementedRegistryServer
}

func (s *Server) ArtifactVerify(
	ctx context.Context,
	in *distribution.ArtifactVerifyRequest,
) (*distribution.ArtifactVerifyResponse, error) {
	return &distribution.ArtifactVerifyResponse{
		Result: distribution.ArtifactVerifyResult_SUCCESS,
	}, nil
}
