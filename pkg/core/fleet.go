package core

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type FleetClientConfig struct {
	ClientSet *kubernetes.Clientset
}

func NewFleetClient(config FleetClientConfig) *FleetClient {
	return &FleetClient{
		clientSet: config.ClientSet,
	}
}

type FleetClient struct {
	clientSet *kubernetes.Clientset
}

func (s *FleetClient) ListWorlds() ([]World, error) {
	nss, err := s.clientSet.
		CoreV1().
		Namespaces().
		List(context.Background(), metav1.ListOptions{
			LabelSelector: "fleet.elkia.io/world=true",
		})
	if err != nil {
		return nil, err
	}
	worlds := make([]World, len(nss.Items))
	for i, ns := range nss.Items {
		gateways, err := s.ListGateways(ns.Name)
		if err != nil {
			return nil, err
		}
		worlds[i] = World{
			ID:       ns.Name,
			Name:     ns.Labels["fleet.elkia.io/world.name"],
			Gateways: gateways,
		}
	}
	return worlds, nil
}

func (s *FleetClient) ListGateways(worldId string) ([]Gateway, error) {
	svcs, err := s.clientSet.
		CoreV1().
		Services(worldId).
		List(context.Background(), metav1.ListOptions{
			LabelSelector: "fleet.elkia.io/gateway=true",
			FieldSelector: "spec.type=LoadBalancer",
		})
	if err != nil {
		return nil, err
	}
	gateways := make([]Gateway, len(svcs.Items))
	for i, svc := range svcs.Items {
		addr, err := s.getGatewayAddrFromService(&svc)
		if err != nil {
			return nil, err
		}
		gateways[i] = Gateway{
			ID:   svc.Name,
			Addr: addr,
		}
	}
	return gateways, nil
}

func (s *FleetClient) getGatewayAddrFromService(svc *corev1.Service) (string, error) {
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

func (s *FleetClient) getGatewayIpFromService(svc *corev1.Service) (string, error) {
	if svc.Spec.LoadBalancerIP == "" {
		return "", fmt.Errorf("service %s/%s has no LoadBalancerIP", svc.Namespace, svc.Name)
	}
	return svc.Spec.LoadBalancerIP, nil
}

func (s *FleetClient) getGatewayPortFromService(svc *corev1.Service) (int32, error) {
	if len(svc.Spec.Ports) == 0 {
		return 0, fmt.Errorf("service %s/%s has no ports", svc.Namespace, svc.Name)
	}
	for _, port := range svc.Spec.Ports {
		if port.Name == "nostale" {
			return port.Port, nil
		}
	}
	return 0, fmt.Errorf("service %s/%s has no port named 'nostale'", svc.Namespace, svc.Name)
}

type World struct {
	ID       string
	Name     string
	Gateways []Gateway
}

type Gateway struct {
	ID         string
	Addr       string
	Population uint
	Capacity   uint
}
