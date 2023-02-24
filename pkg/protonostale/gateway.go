package protonostale

import (
	"bufio"
	"bytes"
	"io"

	eventing "github.com/infinity-blackhole/elkia/pkg/api/eventing/v1alpha1"
	"github.com/infinity-blackhole/elkia/pkg/nostale/monoalphabetic"
	"github.com/sirupsen/logrus"
)

func NewGatewayHandshakeEventReader(r io.Reader) *GatewayHandshakeEventReader {
	return &GatewayHandshakeEventReader{
		EventReader{
			r: bufio.NewReader(r),
		},
	}
}

type GatewayHandshakeEventReader struct {
	EventReader
}

func (r *GatewayHandshakeEventReader) ReadSyncEvent() (*eventing.SyncEvent, error) {
	sn, err := r.ReadSequence()
	if err != nil {
		return nil, err
	}
	return &eventing.SyncEvent{Sequence: sn}, nil
}

func (r *GatewayHandshakeEventReader) ReadAuthHandoffEvent() (*eventing.AuthHandoffEvent, error) {
	keyMsg, err := r.ReadAuthHandoffKeyEvent()
	if err != nil {
		return nil, err
	}
	pwdMsg, err := r.ReadAuthHandoffPasswordEvent()
	if err != nil {
		return nil, err
	}
	return &eventing.AuthHandoffEvent{
		KeyEvent:      keyMsg,
		PasswordEvent: pwdMsg,
	}, nil
}

func (r *GatewayHandshakeEventReader) ReadAuthHandoffKeyEvent() (*eventing.AuthHandoffKeyEvent, error) {
	sn, err := r.ReadSequence()
	if err != nil {
		return nil, err
	}
	key, err := r.ReadUint32()
	if err != nil {
		return nil, err
	}
	return &eventing.AuthHandoffKeyEvent{
		Sequence: sn,
		Key:      key,
	}, nil
}

func (r *GatewayHandshakeEventReader) ReadAuthHandoffPasswordEvent() (*eventing.AuthHandoffPasswordEvent, error) {
	sn, err := r.ReadUint32()
	if err != nil {
		return nil, err
	}
	password, err := r.ReadString()
	if err != nil {
		return nil, err
	}
	return &eventing.AuthHandoffPasswordEvent{
		Sequence: sn,
		Password: password,
	}, nil
}

func NewGatewayChannelEventReader(r io.Reader) *GatewayChannelEventReader {
	return &GatewayChannelEventReader{
		EventReader{
			r: bufio.NewReader(r),
		},
	}
}

type GatewayChannelEventReader struct {
	EventReader
}

func (r *GatewayChannelEventReader) ReadChannelEvent() (*eventing.ChannelEvent, error) {
	sn, err := r.ReadUint32()
	if err != nil {
		return nil, err
	}
	opcode, err := r.ReadOpCode()
	if err != nil {
		return nil, err
	}
	payload, err := r.ReadPayload()
	if err != nil {
		return nil, err
	}
	return &eventing.ChannelEvent{
		Sequence: sn,
		OpCode:   opcode,
		Payload:  &eventing.ChannelEvent_UnknownPayload{UnknownPayload: payload},
	}, nil
}

func NewGatewayWriter(w *bufio.Writer) *GatewayWriter {
	return &GatewayWriter{
		EventWriter: EventWriter{
			w: bufio.NewWriter(monoalphabetic.NewWriter(w)),
		},
	}
}

type GatewayWriter struct {
	EventWriter
}

func NewGatewayHandshakeReader(r *bufio.Reader) *GatewayHandshakeReader {
	return &GatewayHandshakeReader{
		r: monoalphabetic.NewReader(r),
	}
}

type GatewayHandshakeReader struct {
	r *monoalphabetic.Reader
}

func (r *GatewayHandshakeReader) ReadMessageSlice() ([]*GatewayHandshakeEventReader, error) {
	buff, err := r.r.ReadMessageSliceBytes()
	if err != nil {
		return nil, err
	}
	readers := make([]*GatewayHandshakeEventReader, len(buff))
	for i, b := range buff {
		logrus.Debugf("gateway: read %v message", string(b))
		readers[i] = NewGatewayHandshakeEventReader(bytes.NewReader(b))
	}
	return readers, nil
}

func NewGatewayChannelReader(r *bufio.Reader, key uint32) *GatewayChannelReader {
	return &GatewayChannelReader{
		r:   monoalphabetic.NewReaderWithKey(r, key),
		key: key,
	}
}

type GatewayChannelReader struct {
	r   *monoalphabetic.Reader
	key uint32
}

func (r *GatewayChannelReader) ReadMessageSlice() ([]*GatewayChannelEventReader, error) {
	buff, err := r.r.ReadMessageSliceBytes()
	if err != nil {
		return nil, err
	}
	readers := make([]*GatewayChannelEventReader, len(buff))
	for i, b := range buff {
		logrus.Debugf("gateway: read %v message", string(b))
		readers[i] = NewGatewayChannelEventReader(bytes.NewReader(b))
	}
	return readers, nil
}
