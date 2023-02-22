package fleet

import (
	"context"
	"encoding/gob"
	"hash/fnv"

	fleet "github.com/infinity-blackhole/elkia/pkg/api/fleet/v1alpha1"
	"github.com/sirupsen/logrus"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

type FleetServerConfig struct {
	Orchestrator     *Orchestrator
	SessionStore     *SessionStore
	IdentityProvider *IdentityProvider
}

func NewFleetServer(config FleetServerConfig) *FleetServer {
	return &FleetServer{
		orchestrator:     config.Orchestrator,
		sessionStore:     config.SessionStore,
		identityProvider: config.IdentityProvider,
	}
}

type FleetServer struct {
	fleet.UnimplementedFleetServer
	orchestrator     *Orchestrator
	sessionStore     *SessionStore
	identityProvider *IdentityProvider
}

func (s *FleetServer) GetCluster(
	ctx context.Context,
	in *fleet.GetClusterRequest,
) (*fleet.Cluster, error) {
	return s.orchestrator.GetCluster(ctx, in)
}

func (s *FleetServer) ListClusters(
	ctx context.Context,
	in *fleet.ListClusterRequest,
) (*fleet.ListClusterResponse, error) {
	return s.orchestrator.ListClusters(ctx, in)
}

func (s *FleetServer) GetGateway(
	ctx context.Context,
	in *fleet.GetGatewayRequest,
) (*fleet.Gateway, error) {
	return s.orchestrator.GetGateway(ctx, in)
}

func (s *FleetServer) ListGateways(
	ctx context.Context,
	in *fleet.ListGatewayRequest,
) (*fleet.ListGatewayResponse, error) {
	return s.orchestrator.ListGateways(ctx, in)
}

func (s *FleetServer) CreateHandoff(
	ctx context.Context,
	in *fleet.CreateHandoffRequest,
) (*fleet.CreateHandoffResponse, error) {
	session, err := s.identityProvider.
		PerformAuthLoginFlowWithPasswordMethod(
			ctx,
			in.Identifier,
			in.Token,
		)
	if err != nil {
		return nil, err
	}
	logrus.Debugf("fleet: created session: %v", session)
	h := fnv.New32a()
	if err := gob.
		NewEncoder(h).
		Encode(session.Id); err != nil {
		logrus.Fatal(err)
	}
	key := h.Sum32()
	logrus.Debugf("fleet: created key: %v", key)
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
	logrus.Debugf("fleet: created handoff: %v", key)
	return &fleet.CreateHandoffResponse{
		Key: key,
	}, nil
}

func (s *FleetServer) PerformHandoff(
	ctx context.Context,
	in *fleet.PerformHandoffRequest,
) (*emptypb.Empty, error) {
	handoff, err := s.sessionStore.GetHandoffSession(ctx, in.Key)
	if err != nil {
		return nil, err
	}
	logrus.Debugf("fleet: got handoff: %v", handoff)
	keySession, err := s.identityProvider.GetSession(ctx, handoff.Token)
	if err != nil {
		return nil, err
	}
	logrus.Debugf("fleet: got session from key: %v", keySession)
	refreshSession, err := s.identityProvider.
		PerformGatewayLoginFlowWithPasswordMethod(
			ctx,
			handoff.Identifier,
			in.Token,
			handoff.Token,
		)
	if err != nil {
		return nil, err
	}
	if err := s.sessionStore.SetHandoffSession(
		ctx,
		in.Key,
		&fleet.Handoff{
			Id:         refreshSession.Id,
			Identifier: handoff.Identifier,
			Token:      refreshSession.Token,
		},
	); err != nil {
		return nil, err
	}
	logrus.Debugf("fleet: updated handoff: %v", in.Key)
	return &emptypb.Empty{}, nil
}
