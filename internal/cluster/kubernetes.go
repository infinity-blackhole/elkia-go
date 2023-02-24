package cluster

import (
	"context"
	"fmt"
	"net"
	"strconv"

	fleet "github.com/infinity-blackhole/elkia/pkg/api/fleet/v1alpha1"
	"github.com/sirupsen/logrus"
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
	addr, err := s.getGatewayAddrFromService(svc)
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
		Address:    addr,
		Population: uint32(populationUint),
		Capacity:   uint32(capacityUint),
	}, nil
}

func (s *KubernetesClusterServer) getGatewayAddrFromService(
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
	return net.JoinHostPort(ip, port), nil
}

func (s *KubernetesClusterServer) getGatewayIpFromService(
	svc *corev1.Service,
) (string, error) {
	// TODO: We might want to support multiple IPs in the future
	for _, ingress := range svc.Status.LoadBalancer.Ingress {
		if ingress.IP != "" {
			return ingress.IP, nil
		}
		if ingress.Hostname != "" {
			return ingress.Hostname, nil
		}
	}
	return "", fmt.Errorf(
		"service %s/%s has no ingress",
		svc.Namespace,
		svc.Name,
	)
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
