package protonostale

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strconv"

	eventing "github.com/infinity-blackhole/elkia/pkg/api/eventing/v1alpha1"
	"github.com/sirupsen/logrus"
)

var (
	HandoffOpCode = "NoS0575"
)

func ReadAuthLoginEvent(r *bufio.Reader) (*eventing.AuthLoginEvent, error) {
	_, err := ReadString(r)
	if err != nil {
		return nil, err
	}
	identifier, err := ReadString(r)
	if err != nil {
		return nil, err
	}
	logrus.Debugf("read identifier %s", identifier)
	pwd, err := ReadPassword(r)
	if err != nil {
		return nil, err
	}
	logrus.Debugf("read password %s", pwd)
	clientVersion, err := ReadVersion(r)
	if err != nil {
		return nil, err
	}
	logrus.Debugf("read client version %s", clientVersion)
	return &eventing.AuthLoginEvent{
		Identifier:    identifier,
		Password:      pwd,
		ClientVersion: clientVersion,
	}, nil
}

func ReadPassword(r *bufio.Reader) (string, error) {
	pwd, err := ReadField(r)
	if err != nil {
		return "", err
	}
	if len(pwd)%2 == 0 {
		pwd = pwd[3:]
	} else {
		pwd = pwd[4:]
	}
	chunks := bytesChunkEvery(pwd, 2)
	pwd = make([]byte, 0, len(chunks))
	for i := 0; i < len(chunks); i++ {
		pwd = append(pwd, chunks[i][0])
	}
	chunks = bytesChunkEvery(pwd, 2)
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

func WriteGatewayListEvent(
	w io.Writer,
	msg *eventing.GatewayListEvent,
) (n int, err error) {
	var b bytes.Buffer
	if _, err := fmt.Fprintf(&b, "NsTeST %d ", msg.Key); err != nil {
		return n, err
	}
	for _, g := range msg.Gateways {
		if _, err := WriteGateway(&b, g); err != nil {
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

func WriteGateway(w io.Writer, msg *eventing.Gateway) (int, error) {
	return fmt.Fprintf(
		w,
		"%s:%s:%d:%d.%d.%s",
		msg.Host,
		msg.Port,
		msg.Weight,
		msg.WorldId,
		msg.ChannelId,
		msg.WorldName,
	)
}
