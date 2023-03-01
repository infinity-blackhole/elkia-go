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

func ParseAuthFrame(b []byte) (*eventing.AuthInteractRequest, error) {
	fields := bytes.SplitN(b, []byte(" "), 2)
	if len(fields) != 2 {
		return nil, fmt.Errorf("invalid auth frame: %s", b)
	}
	opcode := string(fields[0])
	switch opcode {
	case AuthLoginOpCode:
		authLoginFrame, err := DecodeAuthLoginFrame(fields[1])
		if err != nil {
			return nil, err
		}
		return &eventing.AuthInteractRequest{
			Payload: &eventing.AuthInteractRequest_AuthLoginFrame{
				AuthLoginFrame: authLoginFrame,
			},
		}, nil
	default:
		return nil, fmt.Errorf("unknown auth frame: %s", b)
	}
}

func DecodeAuthLoginFrame(s []byte) (*eventing.AuthLoginFrame, error) {
	fields := bytes.Fields(s)
	if len(fields) != 5 {
		return nil, fmt.Errorf("invalid auth login frame: %s", s)
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
	return &eventing.AuthLoginFrame{
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

func WriteEndpointListFrame(
	w io.Writer,
	msg *eventing.EndpointListFrame,
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
