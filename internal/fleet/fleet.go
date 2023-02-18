package fleet

import (
	"context"
	"encoding/gob"
	"errors"
	"hash/fnv"

	fleet "github.com/infinity-blackhole/elkia/pkg/api/fleet/v1alpha1"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

type FleetConfig struct {
	Orchestrator     *Orchestrator
	SessionStore     *SessionStore
	IdentityProvider *IdentityProvider
}

func NewFleet(config FleetConfig) *Fleet {
	return &Fleet{
		orchestrator:     config.Orchestrator,
		sessionStore:     config.SessionStore,
		identityProvider: config.IdentityProvider,
	}
}

type Fleet struct {
	fleet.UnimplementedFleetServer
	orchestrator     *Orchestrator
	sessionStore     *SessionStore
	identityProvider *IdentityProvider
}

func (s *Fleet) GetCluster(
	ctx context.Context,
	in *fleet.GetClusterRequest,
) (*fleet.Cluster, error) {
	return s.orchestrator.GetCluster(ctx, in)
}

func (s *Fleet) ListClusters(
	ctx context.Context,
	in *fleet.ListClusterRequest,
) (*fleet.ListClusterResponse, error) {
	return s.orchestrator.ListClusters(ctx, in)
}

func (s *Fleet) GetGateway(
	ctx context.Context,
	in *fleet.GetGatewayRequest,
) (*fleet.Gateway, error) {
	return s.orchestrator.GetGateway(ctx, in)
}

func (s *Fleet) ListGateways(
	ctx context.Context,
	in *fleet.ListGatewayRequest,
) (*fleet.ListGatewayResponse, error) {
	return s.orchestrator.ListGateways(ctx, in)
}

func (s *Fleet) CreateHandoff(
	ctx context.Context,
	in *fleet.CreateHandoffRequest,
) (*fleet.CreateHandoffResponse, error) {
	session, err := s.identityProvider.
		PerformLoginFlowWithPasswordMethod(
			ctx,
			in.Identifier,
			in.Token,
		)
	if err != nil {
		return nil, err
	}
	h := fnv.New32a()
	if err := gob.
		NewEncoder(h).
		Encode(session.Id); err != nil {
		panic(err)
	}
	key := h.Sum32()
	if err := s.sessionStore.SetHandoffSession(
		ctx,
		key,
		&fleet.Handoff{
			Id:         session.Id,
			Identifier: in.Identifier,
			Token:      session.Token,
		},
	); err != nil {
		return nil, err
	}
	return &fleet.CreateHandoffResponse{
		Key: key,
	}, nil
}

func (s *Fleet) PerformHandoff(
	ctx context.Context,
	in *fleet.PerformHandoffRequest,
) (*emptypb.Empty, error) {
	handoff, err := s.sessionStore.GetHandoffSession(ctx, in.Key)
	if err != nil {
		return nil, err
	}
	if err := s.sessionStore.DeleteHandoffSession(ctx, in.Key); err != nil {
		return nil, err
	}
	keySession, err := s.identityProvider.GetSession(ctx, handoff.Token)
	if err != nil {
		return nil, err
	}
	if err := s.identityProvider.Logout(ctx, handoff.Token); err != nil {
		return nil, err
	}
	credentialsSession, err := s.identityProvider.
		PerformLoginFlowWithPasswordMethod(
			ctx,
			handoff.Identifier,
			in.Token,
		)
	if err != nil {
		return nil, err
	}
	if keySession.IdentityId != credentialsSession.IdentityId {
		return nil, errors.New("invalid credentials")
	}
	return &emptypb.Empty{}, nil
}
