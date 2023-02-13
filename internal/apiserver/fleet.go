package apiserver

import (
	"context"
	"encoding/gob"
	"errors"
	"fmt"
	"hash/fnv"
	"strconv"

	fleetv1alpha1pb "github.com/infinity-blackhole/elkia/pkg/api/fleet/v1alpha1"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
)

type FleetServiceConfig struct {
	KubernetesClientSet *kubernetes.Clientset
	SessionStore        *SessionStore
	IdentityProvider    *IdentityProvider
}

func NewFleetService(config FleetServiceConfig) *FleetService {
	return &FleetService{
		kubernetesClientSet: config.KubernetesClientSet,
		sessionStore:        config.SessionStore,
		identityProvider:    config.IdentityProvider,
	}
}

type FleetService struct {
	fleetv1alpha1pb.UnimplementedFleetServiceServer
	kubernetesClientSet *kubernetes.Clientset
	sessionStore        *SessionStore
	identityProvider    *IdentityProvider
}

func (s *FleetService) GetCluster(
	ctx context.Context,
	in *fleetv1alpha1pb.GetClusterRequest,
) (*fleetv1alpha1pb.Cluster, error) {
	ns, err := s.kubernetesClientSet.
		CoreV1().
		Namespaces().
		Get(ctx, in.Id, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return s.getClusterFromNamespace(ns)
}

func (s *FleetService) ListClusters(
	ctx context.Context,
	in *fleetv1alpha1pb.ListClusterRequest,
) (*fleetv1alpha1pb.ListClusterResponse, error) {
	nss, err := s.kubernetesClientSet.
		CoreV1().
		Namespaces().
		List(
			ctx,
			metav1.ListOptions{
				LabelSelector: labels.SelectorFromSet(map[string]string{
					"fleet.elkia.io/world": "true",
				}).String(),
			},
		)
	if err != nil {
		return nil, err
	}
	clusters := make([]*fleetv1alpha1pb.Cluster, len(nss.Items))
	for i, ns := range nss.Items {
		clusters[i], err = s.getClusterFromNamespace(&ns)
		if err != nil {
			return nil, err
		}
	}
	return &fleetv1alpha1pb.ListClusterResponse{
		Clusters: clusters,
	}, nil
}

func (s *FleetService) getClusterFromNamespace(
	ns *corev1.Namespace,
) (*fleetv1alpha1pb.Cluster, error) {
	idUint, err := strconv.ParseUint(
		ns.Labels["fleet.elkia.io/cluster-world"],
		10, 32,
	)
	if err != nil {
		return nil, err
	}
	return &fleetv1alpha1pb.Cluster{
		Id:      ns.Name,
		WorldId: uint32(idUint),
		Name:    ns.Labels["fleet.elkia.io/cluster-tenant"],
	}, nil
}

func (s *FleetService) GetGateway(
	ctx context.Context,
	in *fleetv1alpha1pb.GetGatewayRequest,
) (*fleetv1alpha1pb.Gateway, error) {
	svc, err := s.kubernetesClientSet.
		CoreV1().
		Services(in.Id).
		Get(ctx, in.Id, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return s.getGatewayFromService(svc)
}

func (s *FleetService) ListGateways(
	ctx context.Context,
	in *fleetv1alpha1pb.ListGatewayRequest,
) (*fleetv1alpha1pb.ListGatewayResponse, error) {
	gateways := []*fleetv1alpha1pb.Gateway{}
	svcs, err := s.kubernetesClientSet.
		CoreV1().
		Services(in.Id).
		List(ctx, metav1.ListOptions{
			LabelSelector: labels.SelectorFromSet(map[string]string{
				"fleet.elkia.io/gateway": "true",
			}).String(),
			FieldSelector: fields.SelectorFromSet(map[string]string{
				"spec.type": "LoadBalancer",
			}).String(),
		})
	if err != nil {
		return nil, err
	}
	for _, svc := range svcs.Items {
		gateway, err := s.getGatewayFromService(&svc)
		if err != nil {
			return nil, err
		}
		gateways = append(gateways, gateway)
	}
	return &fleetv1alpha1pb.ListGatewayResponse{
		Gateways: gateways,
	}, nil
}

func (s *FleetService) getGatewayFromService(
	svc *corev1.Service,
) (*fleetv1alpha1pb.Gateway, error) {
	addr, err := s.getGatewayAddrFromService(svc)
	if err != nil {
		return nil, err
	}
	idUint, err := strconv.ParseUint(
		svc.Labels["fleet.elkia.io/gateway-channel"],
		10, 32,
	)
	if err != nil {
		return nil, err
	}
	populationUint, err := strconv.ParseUint(
		svc.Labels["fleet.elkia.io/gateway-population"],
		10, 32,
	)
	if err != nil {
		return nil, err
	}
	capacityUint, err := strconv.ParseUint(
		svc.Labels["fleet.elkia.io/gateway-capacity"],
		10, 32,
	)
	if err != nil {
		return nil, err
	}
	return &fleetv1alpha1pb.Gateway{
		Id:         svc.Name,
		ChannelId:  uint32(idUint),
		Address:    addr,
		Population: uint32(populationUint),
		Capacity:   uint32(capacityUint),
	}, nil
}

func (s *FleetService) getGatewayAddrFromService(
	svc *corev1.Service,
) (string, error) {
	ip, err := s.getGatewayIpFromService(svc)
	if err != nil {
		return "", err
	}
	port, err := s.getGatewayPortFromService(svc)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(ip, port), nil
}

func (s *FleetService) getGatewayIpFromService(
	svc *corev1.Service,
) (string, error) {
	if svc.Spec.LoadBalancerIP == "" {
		return "", fmt.Errorf(
			"service %s/%s has no LoadBalancerIP",
			svc.Namespace,
			svc.Name,
		)
	}
	return svc.Spec.LoadBalancerIP, nil
}

func (s *FleetService) getGatewayPortFromService(
	svc *corev1.Service,
) (int32, error) {
	if len(svc.Spec.Ports) == 0 {
		return 0, fmt.Errorf(
			"service %s/%s has no ports",
			svc.Namespace,
			svc.Name,
		)
	}
	for _, port := range svc.Spec.Ports {
		if port.Name == "elkia" {
			return port.Port, nil
		}
	}
	return 0, fmt.Errorf(
		"service %s/%s has no port named 'elkia'",
		svc.Namespace,
		svc.Name,
	)
}

func (s *FleetService) CreateHandoff(
	ctx context.Context,
	in *fleetv1alpha1pb.CreateHandoffRequest,
) (*fleetv1alpha1pb.CreateHandoffResponse, error) {
	session, token, err := s.identityProvider.
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
			Token:      token,
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
	credentialsSession, _, err := s.identityProvider.
		PerformLoginFlowWithPasswordMethod(
			ctx,
			handoff.Identifier,
			in.Token,
		)
	if err != nil {
		return nil, err
	}
	if keySession.Identity.Id != credentialsSession.Identity.Id {
		return nil, errors.New("invalid credentials")
	}
	return &emptypb.Empty{}, nil
}
