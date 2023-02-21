package fleet

import (
	"context"
	"fmt"
	"strconv"

	fleet "github.com/infinity-blackhole/elkia/pkg/api/fleet/v1alpha1"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
)

type OrchestratorConfig struct {
	KubernetesClientSet *kubernetes.Clientset
}

func NewOrchestrator(config OrchestratorConfig) *Orchestrator {
	return &Orchestrator{
		kubernetesClientSet: config.KubernetesClientSet,
	}
}

type Orchestrator struct {
	kubernetesClientSet *kubernetes.Clientset
}

func (s *Orchestrator) GetCluster(
	ctx context.Context,
	in *fleet.GetClusterRequest,
) (*fleet.Cluster, error) {
	ns, err := s.kubernetesClientSet.
		CoreV1().
		Namespaces().
		Get(ctx, in.Id, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return s.getClusterFromNamespace(ns)
}

func (s *Orchestrator) ListClusters(
	ctx context.Context,
	in *fleet.ListClusterRequest,
) (*fleet.ListClusterResponse, error) {
	nss, err := s.kubernetesClientSet.
		CoreV1().
		Namespaces().
		List(
			ctx,
			metav1.ListOptions{
				LabelSelector: labels.SelectorFromSet(map[string]string{
					"fleet.elkia.io/cluster": "true",
				}).String(),
			},
		)
	if err != nil {
		return nil, err
	}
	logrus.Debugf("fleet: found %d clusters", len(nss.Items))
	clusters := make([]*fleet.Cluster, len(nss.Items))
	for i, ns := range nss.Items {
		clusters[i], err = s.getClusterFromNamespace(&ns)
		if err != nil {
			return nil, err
		}
	}
	return &fleet.ListClusterResponse{
		Clusters: clusters,
	}, nil
}

func (s *Orchestrator) getClusterFromNamespace(
	ns *corev1.Namespace,
) (*fleet.Cluster, error) {
	idUint, err := strconv.ParseUint(
		ns.Labels["fleet.elkia.io/world"],
		10, 32,
	)
	if err != nil {
		return nil, err
	}
	return &fleet.Cluster{
		Id:      ns.Name,
		WorldId: uint32(idUint),
		Name:    ns.Labels["fleet.elkia.io/instance"],
	}, nil
}

func (s *Orchestrator) GetGateway(
	ctx context.Context,
	in *fleet.GetGatewayRequest,
) (*fleet.Gateway, error) {
	svc, err := s.kubernetesClientSet.
		CoreV1().
		Services(in.Id).
		Get(ctx, in.Id, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	logrus.Debugf("fleet: found gateway %s", svc.Name)
	return s.getGatewayFromService(svc)
}

func (s *Orchestrator) ListGateways(
	ctx context.Context,
	in *fleet.ListGatewayRequest,
) (*fleet.ListGatewayResponse, error) {
	gateways := []*fleet.Gateway{}
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
	logrus.Debugf("fleet: found %d gateways", len(svcs.Items))
	for _, svc := range svcs.Items {
		gateway, err := s.getGatewayFromService(&svc)
		if err != nil {
			return nil, err
		}
		gateways = append(gateways, gateway)
	}
	return &fleet.ListGatewayResponse{
		Gateways: gateways,
	}, nil
}

func (s *Orchestrator) getGatewayFromService(
	svc *corev1.Service,
) (*fleet.Gateway, error) {
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
	return &fleet.Gateway{
		Id:         svc.Name,
		ChannelId:  uint32(idUint),
		Address:    addr,
		Population: uint32(populationUint),
		Capacity:   uint32(capacityUint),
	}, nil
}

func (s *Orchestrator) getGatewayAddrFromService(
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

func (s *Orchestrator) getGatewayIpFromService(
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

func (s *Orchestrator) getGatewayPortFromService(
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
