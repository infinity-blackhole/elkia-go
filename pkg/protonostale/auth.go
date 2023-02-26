package protonostale

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"

	eventing "github.com/infinity-blackhole/elkia/pkg/api/eventing/v1alpha1"
)

var (
	HandoffOpCode = "NoS0575"
)

func ParseAuthEvent(s string) (*eventing.AuthInteractRequest, error) {
	fields := strings.Fields(s)
	if len(fields) != 2 {
		return nil, fmt.Errorf("invalid auth event: %s", s)
	}
	switch fields[1] {
	case HandoffOpCode:
		authLoginEvent, err := ParseAuthLoginEvent(s)
		if err != nil {
			return nil, err
		}
		return &eventing.AuthInteractRequest{
			Payload: &eventing.AuthInteractRequest_AuthLoginEvent{
				AuthLoginEvent: authLoginEvent,
			},
		}, nil
	default:
		return nil, fmt.Errorf("unknown auth event: %s", s)
	}
}

func ParseAuthLoginEvent(s string) (*eventing.AuthLoginEvent, error) {
	fields := strings.Fields(s)
	if len(fields) != 5 {
		return nil, fmt.Errorf("invalid auth login event: %s", s)
	}
	identifier := fields[1]
	pwd, err := ParsePassword(fields[2])
	if err != nil {
		return nil, err
	}
	clientVersion, err := ParseVersion(fields[4])
	if err != nil {
		return nil, err
	}
	return &eventing.AuthLoginEvent{
		Identifier:    identifier,
		Password:      pwd,
		ClientVersion: clientVersion,
	}, nil
}

func ParsePassword(s string) (string, error) {
	b := []byte(s)
	if len(b)%2 == 0 {
		b = b[3:]
	} else {
		b = b[4:]
	}
	chunks := bytesChunkEvery(b, 2)
	b = make([]byte, 0, len(chunks))
	for i := 0; i < len(chunks); i++ {
		b = append(b, chunks[i][0])
	}
	chunks = bytesChunkEvery(b, 2)
	var result []byte
	for _, chunk := range chunks {
		value, err := strconv.ParseInt(string(chunk), 16, 64)
		if err != nil {
			return "", err
		}
		result = append(result, byte(value))
	}
	return string(result), nil
}

func WriteEndpointListEvent(
	w io.Writer,
	msg *eventing.EndpointListEvent,
) (n int, err error) {
	var b bytes.Buffer
	if _, err := fmt.Fprintf(&b, "NsTeST %d ", msg.Key); err != nil {
		return n, err
	}
	for _, m := range msg.Endpoints {
		if _, err := WriteEndpoint(&b, m); err != nil {
			return n, err
		}
		if err := b.WriteByte(' '); err != nil {
			return n, err
		}
	}
	_, err = b.WriteString("-1:-1:-1:10000.10000.1")
	if err != nil {
		return n, err
	}
	return w.Write(b.Bytes())
}

func WriteEndpoint(w io.Writer, m *eventing.Endpoint) (int, error) {
	return fmt.Fprintf(
		w,
		"%s:%s:%d:%d.%d.%s",
		m.Host,
		m.Port,
		m.Weight,
		m.WorldId,
		m.ChannelId,
		m.WorldName,
	)
}
