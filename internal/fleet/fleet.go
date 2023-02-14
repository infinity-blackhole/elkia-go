package fleet

import (
	"context"
	"encoding/gob"
	"errors"
	"hash/fnv"

	fleetv1alpha1pb "github.com/infinity-blackhole/elkia/pkg/api/fleet/v1alpha1"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

type FleetServiceConfig struct {
	Orchestrator     *Orchestrator
	SessionStore     *SessionStore
	IdentityProvider *IdentityProvider
}

func NewFleetService(config FleetServiceConfig) *FleetService {
	return &FleetService{
		orchestrator:     config.Orchestrator,
		sessionStore:     config.SessionStore,
		identityProvider: config.IdentityProvider,
	}
}

type FleetService struct {
	fleetv1alpha1pb.UnimplementedFleetServiceServer
	orchestrator     *Orchestrator
	sessionStore     *SessionStore
	identityProvider *IdentityProvider
}

func (s *FleetService) GetCluster(
	ctx context.Context,
	in *fleetv1alpha1pb.GetClusterRequest,
) (*fleetv1alpha1pb.Cluster, error) {
	return s.orchestrator.GetCluster(ctx, in)
}

func (s *FleetService) ListClusters(
	ctx context.Context,
	in *fleetv1alpha1pb.ListClusterRequest,
) (*fleetv1alpha1pb.ListClusterResponse, error) {
	return s.orchestrator.ListClusters(ctx, in)
}

func (s *FleetService) GetGateway(
	ctx context.Context,
	in *fleetv1alpha1pb.GetGatewayRequest,
) (*fleetv1alpha1pb.Gateway, error) {
	return s.orchestrator.GetGateway(ctx, in)
}

func (s *FleetService) ListGateways(
	ctx context.Context,
	in *fleetv1alpha1pb.ListGatewayRequest,
) (*fleetv1alpha1pb.ListGatewayResponse, error) {
	return s.orchestrator.ListGateways(ctx, in)
}

func (s *FleetService) CreateHandoff(
	ctx context.Context,
	in *fleetv1alpha1pb.CreateHandoffRequest,
) (*fleetv1alpha1pb.CreateHandoffResponse, error) {
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
		&fleetv1alpha1pb.Handoff{
			Id:         session.Id,
			Identifier: in.Identifier,
			Token:      session.Token,
		},
	); err != nil {
		return nil, err
	}
	return &fleetv1alpha1pb.CreateHandoffResponse{
		Key: key,
	}, nil
}

func (s *FleetService) PerformHandoff(
	ctx context.Context,
	in *fleetv1alpha1pb.PerformHandoffRequest,
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
