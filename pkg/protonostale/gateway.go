package protonostale

import (
	"bufio"
	"bytes"
	"io"

	eventing "github.com/infinity-blackhole/elkia/pkg/api/eventing/v1alpha1"
	"github.com/infinity-blackhole/elkia/pkg/nostale/monoalphabetic"
	"github.com/sirupsen/logrus"
)

func NewGatewayMessageReader(r io.Reader) *GatewayMessageReader {
	return &GatewayMessageReader{
		MessageReader{
			r: NewFieldReader(r),
		},
	}
}

type GatewayMessageReader struct {
	MessageReader
}

func (r *GatewayMessageReader) ReadSyncMessage() (*eventing.ChannelMessage, error) {
	sn, err := r.r.ReadUint32()
	if err != nil {
		return nil, err
	}
	return &eventing.ChannelMessage{
		Sequence: sn,
		Payload:  &eventing.ChannelMessage_SyncMessage{SyncMessage: &eventing.SyncMessage{}},
	}, nil
}

func (r *GatewayMessageReader) ReadChannelMessage() (*eventing.ChannelMessage, error) {
	sn, err := r.r.ReadUint32()
	if err != nil {
		return nil, err
	}
	payload, err := r.r.ReadPayload()
	if err != nil {
		return nil, err
	}
	return &eventing.ChannelMessage{
		Sequence: sn,
		Payload:  &eventing.ChannelMessage_UnknownMessage{UnknownMessage: payload},
	}, nil
}

func (r *GatewayMessageReader) ReadAuthHandoffMessage() (*eventing.AuthHandoffMessage, error) {
	keyMsg, err := r.ReadKeyMessage()
	if err != nil {
		return nil, err
	}
	pwdMsg, err := r.ReadPasswordMessage()
	if err != nil {
		return nil, err
	}
	return &eventing.AuthHandoffMessage{
		KeyMessage:      keyMsg,
		PasswordMessage: pwdMsg,
	}, nil
}

func (r *GatewayMessageReader) ReadKeyMessage() (*eventing.AuthHandoffKeyMessage, error) {
	sn, err := r.r.ReadUint32()
	if err != nil {
		return nil, err
	}
	key, err := r.r.ReadUint32()
	if err != nil {
		return nil, err
	}
	return &eventing.AuthHandoffKeyMessage{
		Sequence: sn,
		Key:      key,
	}, nil
}

func (r *GatewayMessageReader) ReadPasswordMessage() (*eventing.AuthHandoffPasswordMessage, error) {
	sn, err := r.r.ReadUint32()
	if err != nil {
		return nil, err
	}
	password, err := r.r.ReadString()
	if err != nil {
		return nil, err
	}
	return &eventing.AuthHandoffPasswordMessage{
		Sequence: sn,
		Password: password,
	}, nil
}

func NewGatewayWriter(w *bufio.Writer) *GatewayWriter {
	return &GatewayWriter{
		Writer: Writer{
			w: bufio.NewWriter(monoalphabetic.NewWriter(w)),
		},
	}
}

type GatewayWriter struct {
	Writer
}

func NewGatewayReader(r *bufio.Reader) *GatewayReader {
	return &GatewayReader{
		r: monoalphabetic.NewReader(r),
	}
}

type GatewayReader struct {
	r *monoalphabetic.Reader
}

func (r *GatewayReader) SetKey(key uint32) {
	r.r.SetKey(key)
}

func (r *GatewayReader) ReadMessageSlice() ([]*GatewayMessageReader, error) {
	buff, err := r.r.ReadMessageSliceBytes()
	if err != nil {
		return nil, err
	}
	readers := make([]*GatewayMessageReader, len(buff))
	for i, b := range buff {
		logrus.Debugf("gateway: read %v message", string(b))
		readers[i] = NewGatewayMessageReader(bytes.NewReader(b))
	}
	return readers, nil
}
