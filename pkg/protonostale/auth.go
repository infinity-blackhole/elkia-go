package protonostale

import (
	"bytes"
	"fmt"
	"io"
	"strconv"

	eventing "github.com/infinity-blackhole/elkia/pkg/api/eventing/v1alpha1"
)

var (
	AuthLoginOpCode   = "NoS0575"
	DialogErrorOpCode = "failc"
	DialogInfoOpCode  = "info"
)

func ParseAuthEvent(b []byte) (*eventing.AuthInteractRequest, error) {
	fields := bytes.SplitN(b, []byte(" "), 2)
	if len(fields) != 2 {
		return nil, fmt.Errorf("invalid auth event: %s", b)
	}
	opcode := string(fields[0])
	switch opcode {
	case AuthLoginOpCode:
		authLoginEvent, err := DecodeAuthLoginEvent(fields[1])
		if err != nil {
			return nil, err
		}
		return &eventing.AuthInteractRequest{
			Payload: &eventing.AuthInteractRequest_AuthLoginEvent{
				AuthLoginEvent: authLoginEvent,
			},
		}, nil
	default:
		return nil, fmt.Errorf("unknown auth event: %s", b)
	}
}

func DecodeAuthLoginEvent(s []byte) (*eventing.AuthLoginEvent, error) {
	fields := bytes.Fields(s)
	if len(fields) != 5 {
		return nil, fmt.Errorf("invalid auth login event: %s", s)
	}
	identifier := string(fields[1])
	pwd, err := ParsePassword(fields[2])
	if err != nil {
		return nil, err
	}
	clientVersion, err := DecodeVersion(fields[4])
	if err != nil {
		return nil, err
	}
	return &eventing.AuthLoginEvent{
		Identifier:    identifier,
		Password:      pwd,
		ClientVersion: clientVersion,
	}, nil
}

func ParsePassword(s []byte) (string, error) {
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
	if _, err := fmt.Fprintf(&b, "NsTeST %d ", msg.Code); err != nil {
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
	return fmt.Fprintln(w, b.String())
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
