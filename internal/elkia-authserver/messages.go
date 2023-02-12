package elkiaauthserver

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"net"
	"strings"
)

type CredentialsMessage struct {
	Identifier     string
	Password       string
	ClientVersion  string
	ClientChecksum string
}

func ParseCredentialsMessage(msg []byte) (*CredentialsMessage, error) {
	ss := bytes.Split(msg, []byte(" "))
	if len(ss) != 7 {
		return nil, errors.New("invalid auth message")
	}
	return &CredentialsMessage{
		Identifier:     string(ss[1]),
		Password:       string(ss[2]),
		ClientVersion:  string(ss[4]),
		ClientChecksum: string(ss[6]),
	}, nil
}

var GatewayTerminator = "-1 -1 -1 10000 10000 1"

type ListGatewaysMessage struct {
	Key      uint32
	Gateways []Gateway
}

func (e ListGatewaysMessage) String() string {
	gateways := make([]string, len(e.Gateways))
	for i, g := range e.Gateways {
		gateways[i] = g.String()
	}
	gateways = append(gateways, GatewayTerminator)
	return fmt.Sprintf(
		"NsTeST %d %s",
		e.Key,
		strings.Join(gateways, " "),
	)
}

type Gateway struct {
	ID         uint32
	Addr       string
	Population uint
	Capacity   uint
	WorldID    uint32
	WorldName  string
}

func (g Gateway) String() string {
	host, port, err := net.SplitHostPort(g.Addr)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf(
		"%s %s %d %.f %d %d %s",
		host,
		port,
		g.Population,
		math.Round(float64(g.Population)/float64(g.Capacity)*20)+1,
		g.WorldID,
		g.ID,
		g.WorldName,
	)
}
