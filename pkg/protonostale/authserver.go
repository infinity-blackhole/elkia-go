package protonostale

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"net"
	"strings"

	eventingv1alpha1pb "github.com/infinity-blackhole/elkia/pkg/api/eventing/v1alpha1"
)

func ParseRequestHandoffMessage(
	msg []byte,
) (*eventingv1alpha1pb.RequestHandoffMessage, error) {
	ss := bytes.Split(msg, []byte(" "))
	if len(ss) != 7 {
		return nil, errors.New("invalid auth message")
	}
	return &eventingv1alpha1pb.RequestHandoffMessage{
		Identifier:     string(ss[1]),
		Password:       string(ss[2]),
		ClientVersion:  string(ss[4]),
		ClientChecksum: string(ss[6]),
	}, nil
}

var GatewayTerminator = "-1 -1 -1 10000 10000 1"

func MarshallProposeHandoffMessage(
	msg *eventingv1alpha1pb.ProposeHandoffMessage,
) string {
	gateways := make([]string, len(msg.Gateways))
	for i, g := range msg.Gateways {
		gateways[i] = MarshallGateway(g)
	}
	gateways = append(gateways, GatewayTerminator)
	return fmt.Sprintf(
		"NsTeST %d %s",
		msg.Key,
		strings.Join(gateways, " "),
	)
}

func MarshallGateway(
	msg *eventingv1alpha1pb.Gateway,
) string {
	host, port, err := net.SplitHostPort(msg.Address)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf(
		"%s %s %d %.f %d %d %s",
		host,
		port,
		msg.Population,
		math.Round(float64(msg.Population)/float64(msg.Capacity)*20)+1,
		msg.WorldId,
		msg.Id,
		msg.WorldName,
	)
}
