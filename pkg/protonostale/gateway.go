package protonostale

import (
	"bufio"
	"bytes"
	"io"

	eventing "github.com/infinity-blackhole/elkia/pkg/api/eventing/v1alpha1"
	"github.com/infinity-blackhole/elkia/pkg/nostale/monoalphabetic"
)

func NewGatewayMessageReader(r io.Reader) *GatewayMessageReader {
	return &GatewayMessageReader{
		r: NewFieldReader(bufio.NewReader(r)),
	}
}

type GatewayMessageReader struct {
	r *FieldReader
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

func (r *GatewayMessageReader) ReadPerformHandoffMessage() (*eventing.PerformHandoffMessage, error) {
	keyMsg, err := r.ReadKeyMessage()
	if err != nil {
		return nil, err
	}
	pwdMsg, err := r.ReadPasswordMessage()
	if err != nil {
		return nil, err
	}
	return &eventing.PerformHandoffMessage{
		KeyMessage:      keyMsg,
		PasswordMessage: pwdMsg,
	}, nil
}

func (r *GatewayMessageReader) ReadKeyMessage() (*eventing.PerformHandoffKeyMessage, error) {
	sn, err := r.r.ReadUint32()
	if err != nil {
		return nil, err
	}
	key, err := r.r.ReadUint32()
	if err != nil {
		return nil, err
	}
	return &eventing.PerformHandoffKeyMessage{
		Sequence: sn,
		Key:      key,
	}, nil
}

func (r *GatewayMessageReader) ReadPasswordMessage() (*eventing.PerformHandoffPasswordMessage, error) {
	sn, err := r.r.ReadUint32()
	if err != nil {
		return nil, err
	}
	password, err := r.r.ReadString()
	if err != nil {
		return nil, err
	}
	return &eventing.PerformHandoffPasswordMessage{
		Sequence: sn,
		Password: password,
	}, nil
}

func NewGatewayWriter(w *bufio.Writer) *GatewayWriter {
	return &GatewayWriter{
		w: monoalphabetic.NewWriter(w),
	}
}

type GatewayWriter struct {
	w *monoalphabetic.Writer
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
		readers[i] = NewGatewayMessageReader(bytes.NewReader(b))
	}
	return readers, nil
}
