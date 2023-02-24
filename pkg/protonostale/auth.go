package protonostale

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"math"
	"strconv"

	eventing "github.com/infinity-blackhole/elkia/pkg/api/eventing/v1alpha1"
	"github.com/infinity-blackhole/elkia/pkg/nostale/simplesubtitution"
	"github.com/sirupsen/logrus"
)

var (
	HandoffOpCode = "NoS0575"
)

func NewAuthEventReader(r io.Reader) *AuthEventReader {
	return &AuthEventReader{
		EventReader: EventReader{
			r: bufio.NewReader(r),
		},
	}
}

type AuthEventReader struct {
	EventReader
}

func (r *AuthEventReader) ReadAuthLoginEvent() (*eventing.AuthLoginEvent, error) {
	logrus.Debugf("reading request handoff message")
	_, err := r.ReadString()
	if err != nil {
		return nil, err
	}
	logrus.Debugf("reading identifier")
	identifier, err := r.ReadString()
	if err != nil {
		return nil, err
	}
	logrus.Debugf("reading password")
	pwd, err := r.ReadPassword()
	if err != nil {
		return nil, err
	}
	logrus.Debugf("reading client version")
	clientVersion, err := r.ReadVersion()
	if err != nil {
		return nil, err
	}
	return &eventing.AuthLoginEvent{
		Identifier:    identifier,
		Password:      pwd,
		ClientVersion: clientVersion,
	}, nil
}

func (r *AuthEventReader) ReadPassword() (string, error) {
	pwd, err := r.ReadField()
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

func (r *AuthEventReader) ReadVersion() (string, error) {
	version, err := r.ReadField()
	if err != nil {
		return "", err
	}
	return NewVersionReader(bytes.NewReader(version)).ReadVersion()
}

func NewAuthReader(r *bufio.Reader) *AuthReader {
	return &AuthReader{
		r: simplesubtitution.NewReader(r),
	}
}

type AuthReader struct {
	r *simplesubtitution.Reader
}

func (r *AuthReader) ReadMessage() (*AuthEventReader, error) {
	buff, err := r.r.ReadMessageBytes()
	if err != nil {
		return nil, err
	}
	logrus.Debugf("auth: read %s messages", string(buff))
	return NewAuthEventReader(bytes.NewReader(buff)), nil
}

func NewAuthWriter(w *bufio.Writer) *AuthWriter {
	return &AuthWriter{
		EventWriter: EventWriter{
			w: bufio.NewWriter(simplesubtitution.NewWriter(w)),
		},
	}
}

type AuthWriter struct {
	EventWriter
}

func (w *AuthWriter) WriteGatewayListEvent(
	msg *eventing.GatewayListEvent,
) error {
	_, err := fmt.Fprintf(w.w, "%s %d", "NsTeST", msg.Key)
	if err != nil {
		return err
	}
	for _, g := range msg.Gateways {
		if err := w.WriteGateway(g); err != nil {
			return err
		}
		if err := w.w.WriteByte(' '); err != nil {
			return err
		}
	}
	if _, err = w.w.WriteString("-1 -1 -1 10000 10000 1"); err != nil {
		return err
	}
	return w.w.Flush()
}

func (w *AuthWriter) WriteGateway(
	msg *eventing.Gateway,
) error {
	_, err := fmt.Fprintf(
		w.w,
		"%s %s %d %.f %d %d %s",
		msg.Host,
		msg.Port,
		msg.Population,
		math.Round(float64(msg.Population)/float64(msg.Capacity)*20)+1,
		msg.WorldId,
		msg.ChannelId,
		msg.WorldName,
	)
	return err
}
