package cluster

import (
	"context"
	"fmt"
	"net"
	"strconv"

	"github.com/sirupsen/logrus"
	fleet "go.shikanime.studio/elkia/pkg/api/fleet/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
)

type KubernetesClusterServerConfig struct {
	Namespace        string
	KubernetesClient *kubernetes.Clientset
}

func NewKubernetesClusterServer(config KubernetesClusterServerConfig) *KubernetesClusterServer {
	return &KubernetesClusterServer{
		namespace: config.Namespace,
		kube:      config.KubernetesClient,
	}
}

type KubernetesClusterServer struct {
	fleet.UnimplementedClusterServer
	namespace string
	kube      *kubernetes.Clientset
}

func (s *KubernetesClusterServer) MemberList(
	ctx context.Context,
	in *fleet.MemberListRequest,
) (*fleet.MemberListResponse, error) {
	svcs, err := s.kube.
		CoreV1().
		Services(s.namespace).
		List(
			ctx,
			metav1.ListOptions{
				LabelSelector: labels.SelectorFromSet(map[string]string{
					"fleet.elkia.io/managed": "true",
				}).String(),
			},
		)
	if err != nil {
		return nil, err
	}
	logrus.Debugf("fleet: found %d members", len(svcs.Items))
	members := make([]*fleet.Member, len(svcs.Items))
	for i, ns := range svcs.Items {
		members[i], err = s.getMemberFromService(&ns)
		if err != nil {
			return nil, err
		}
	}
	return &fleet.MemberListResponse{
		Members: members,
	}, nil
}

func (s *KubernetesClusterServer) getMemberFromService(
	svc *corev1.Service,
) (*fleet.Member, error) {
	worldIdUint, err := strconv.ParseUint(
		svc.Labels["fleet.elkia.io/world-id"],
		10, 32,
	)
	if err != nil {
		return nil, err
	}
	channelIdUint, err := strconv.ParseUint(
		svc.Labels["fleet.elkia.io/channel-id"],
		10, 32,
	)
	if err != nil {
		return nil, err
	}
	addresses, err := s.listGatewayAddrFromService(svc)
	if err != nil {
		return nil, err
	}
	populationUint, err := strconv.ParseUint(
		svc.Labels["fleet.elkia.io/population"],
		10, 32,
	)
	if err != nil {
		return nil, err
	}
	capacityUint, err := strconv.ParseUint(
		svc.Labels["fleet.elkia.io/capacity"],
		10, 32,
	)
	if err != nil {
		return nil, err
	}
	return &fleet.Member{
		Id:         svc.Name,
		WorldId:    uint32(worldIdUint),
		ChannelId:  uint32(channelIdUint),
		Name:       svc.Labels["fleet.elkia.io/world-name"],
		Addresses:  addresses,
		Population: uint32(populationUint),
		Capacity:   uint32(capacityUint),
	}, nil
}

func (s *KubernetesClusterServer) listGatewayAddrFromService(
	svc *corev1.Service,
) ([]string, error) {
	endpoints, err := s.listGatewayIpFromService(svc)
	if err != nil {
		return nil, err
	}
	port, err := s.getGatewayPortFromService(svc)
	if err != nil {
		return nil, err
	}
	addresses := make([]string, len(endpoints))
	for i, endpoint := range endpoints {
		addresses[i] = net.JoinHostPort(endpoint, port)
	}
	return addresses, nil
}

func (s *KubernetesClusterServer) listGatewayIpFromService(
	svc *corev1.Service,
) ([]string, error) {
	endpoints := make([]string, 0)
	for _, ingress := range svc.Status.LoadBalancer.Ingress {
		if ingress.IP != "" {
			endpoints = append(endpoints, ingress.IP)
		} else if ingress.Hostname != "" {
			endpoints = append(endpoints, ingress.Hostname)
		}
	}
	if len(endpoints) == 0 {
		logrus.Warnf(
			"fleet: service %s/%s has no ingress defaulting to 127.0.0.1",
			svc.Namespace,
			svc.Name,
		)
		return []string{"127.0.0.1"}, nil
	}
	return endpoints, nil
}

func (s *KubernetesClusterServer) getGatewayPortFromService(
	svc *corev1.Service,
) (string, error) {
	if len(svc.Spec.Ports) == 0 {
		return "", fmt.Errorf(
			"service %s/%s has no ports",
			svc.Namespace,
			svc.Name,
		)
	}
	for _, port := range svc.Spec.Ports {
		if port.Name == svc.Labels["fleet.elkia.io/port"] {
			return strconv.Itoa(int(port.NodePort)), nil
		}
	}
	return "", fmt.Errorf(
		"service %s/%s has no port named 'elkia'",
		svc.Namespace,
		svc.Name,
	)
}
