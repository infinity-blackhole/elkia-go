package protonostale

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"math"

	eventing "github.com/infinity-blackhole/elkia/pkg/api/eventing/v1alpha1"
	"github.com/infinity-blackhole/elkia/pkg/nostale/simplesubtitution"
)

func NewAuthServerMessageReader(r io.Reader) *AuthServerMessageReader {
	return &AuthServerMessageReader{
		r: NewFieldReader(r),
	}
}

type AuthServerMessageReader struct {
	r *FieldReader
}

func (r *AuthServerMessageReader) ReadRequestHandoffMessage() (*eventing.RequestHandoffMessage, error) {
	_, err := r.r.ReadString()
	if err != nil {
		return nil, err
	}
	identifier, err := r.r.ReadString()
	if err != nil {
		return nil, err
	}
	pwd, err := r.r.ReadString()
	if err != nil {
		return nil, err
	}
	_, err = r.r.ReadString()
	if err != nil {
		return nil, err
	}
	_, err = r.r.ReadString()
	if err != nil {
		return nil, err
	}
	clientVersion, err := r.r.ReadString()
	if err != nil {
		return nil, err
	}
	_, err = r.r.ReadString()
	if err != nil {
		return nil, err
	}
	clientChecksum, err := r.r.ReadString()
	if err != nil {
		return nil, err
	}
	return &eventing.RequestHandoffMessage{
		Identifier:     identifier,
		Password:       pwd,
		ClientVersion:  clientVersion,
		ClientChecksum: clientChecksum,
	}, nil
}

func NewAuthServerReader(r *bufio.Reader) *AuthServerReader {
	return &AuthServerReader{
		r: simplesubtitution.NewReader(r),
	}
}

type AuthServerReader struct {
	r *simplesubtitution.Reader
}

func (r *AuthServerReader) ReadMessage() (*AuthServerMessageReader, error) {
	buff, err := r.r.ReadMessageBytes()
	if err != nil {
		return nil, err
	}
	return NewAuthServerMessageReader(bytes.NewReader(buff)), nil
}

func NewAuthServerWriter(w *bufio.Writer) *AuthServerWriter {
	return &AuthServerWriter{
		ClientWriter: ClientWriter{
			w: bufio.NewWriter(simplesubtitution.NewWriter(w)),
		},
	}
}

type AuthServerWriter struct {
	ClientWriter
}

func (w *AuthServerWriter) WriteProposeHandoffMessage(
	msg *eventing.ProposeHandoffMessage,
) error {
	_, err := fmt.Fprintf(w.w, "%s %d", "NsTeST", msg.Key)
	if err != nil {
		return err
	}
	for _, g := range msg.Gateways {
		if err := w.WriteGatewayMessage(g); err != nil {
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

func (w *AuthServerWriter) WriteGatewayMessage(
	msg *eventing.GatewayMessage,
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
