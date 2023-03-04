package registry

import (
	"context"

	distribution "github.com/infinity-blackhole/elkia/pkg/api/distribution/v1alpha1"
	"golang.org/x/mod/semver"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Artifact struct {
	Version  string
	Checksum string
}

type MemoryServerConfig struct {
	Artifacts []*Artifact
}

func NewMemoryServer(cfg MemoryServerConfig) *MemoryServer {
	return &MemoryServer{}
}

type MemoryServer struct {
	distribution.UnimplementedRegistryServer
	artifacts []*Artifact
}

func (s *MemoryServer) ArtifactVerify(
	ctx context.Context,
	in *distribution.ArtifactVerifyRequest,
) (*distribution.ArtifactVerifyResponse, error) {
	latest, err := s.ArtifactLatest(ctx, &distribution.ArtifactLatestRequest{})
	if err != nil {
		return nil, err
	}
	if latest.Version == in.Version {
		if latest.Checksum == in.Checksum {
			return &distribution.ArtifactVerifyResponse{
				Result: distribution.ArtifactVerifyResult_SUCCESS,
			}, nil
		}
		return &distribution.ArtifactVerifyResponse{
			Result: distribution.ArtifactVerifyResult_CORRUPTED,
		}, nil
	}
	return &distribution.ArtifactVerifyResponse{
		Result: distribution.ArtifactVerifyResult_OUTDATED,
	}, nil
}

func (s *MemoryServer) ArtifactLatest(
	ctx context.Context,
	in *distribution.ArtifactLatestRequest,
) (*distribution.ArtifactLatestResponse, error) {
	versions := make([]string, len(s.artifacts))
	for i, artifact := range s.artifacts {
		switch in.Channel {
		case distribution.Channel_STABLE:
			if semver.Prerelease(artifact.Version) == "" {
				versions[i] = artifact.Version
			}
		default:
			versions[i] = artifact.Version
		}
	}
	semver.Sort(versions)
	var artifact *Artifact
	for _, artifact = range s.artifacts {
		if artifact.Version == versions[len(versions)-1] {
			return &distribution.ArtifactLatestResponse{
				Version:  artifact.Version,
				Checksum: artifact.Checksum,
			}, nil
		}
	}
	return nil, status.Errorf(codes.NotFound, "no artifacts found")
}
