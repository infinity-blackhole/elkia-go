package protonostale

import (
	"bufio"
	"io"
	"strconv"

	eventing "github.com/infinity-blackhole/elkia/pkg/api/eventing/v1alpha1"
)

func NewFieldReader(r *bufio.Reader) *FieldReader {
	return &FieldReader{
		R: r,
	}
}

type FieldReader struct {
	R *bufio.Reader
}

func (r *FieldReader) ReadString() (string, error) {
	rf, err := r.ReadField()
	if err != nil {
		return "", err
	}
	return string(rf), nil
}

func (r *FieldReader) ReadUint32() (uint32, error) {
	rf, err := r.ReadField()
	if err != nil {
		return 0, err
	}
	key, err := strconv.ParseUint(string(rf), 10, 32)
	if err != nil {
		return 0, err
	}
	return uint32(key), nil
}

func (r *FieldReader) ReadField() ([]byte, error) {
	return r.R.ReadBytes(' ')
}

func (r *FieldReader) ReadPayload() ([]byte, error) {
	return r.R.ReadBytes('\n')
}

func NewAuthServerMessageReader(r io.Reader) *AuthServerMessageReader {
	return &AuthServerMessageReader{
		R: NewFieldReader(bufio.NewReader(r)),
	}
}

type AuthServerMessageReader struct {
	R *FieldReader
}

func (r *AuthServerMessageReader) ReadRequestHandoffMessage() (*eventing.RequestHandoffMessage, error) {
	identifier, err := r.R.ReadString()
	if err != nil {
		return nil, err
	}
	pwd, err := r.R.ReadString()
	if err != nil {
		return nil, err
	}
	clientVersion, err := r.R.ReadString()
	if err != nil {
		return nil, err
	}
	clientChecksum, err := r.R.ReadString()
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

func NewGatewayMessageReader(r io.Reader) *GatewayMessageReader {
	return &GatewayMessageReader{
		R: NewFieldReader(bufio.NewReader(r)),
	}
}

type GatewayMessageReader struct {
	R *FieldReader
}

func (r *GatewayMessageReader) ReadSyncMessage() (*eventing.ChannelMessage, error) {
	sn, err := r.R.ReadUint32()
	if err != nil {
		return nil, err
	}
	return &eventing.ChannelMessage{
		Sequence: sn,
		Payload:  &eventing.ChannelMessage_SyncMessage{SyncMessage: &eventing.SyncMessage{}},
	}, nil
}

func (r *GatewayMessageReader) ReadChannelMessage() (*eventing.ChannelMessage, error) {
	sn, err := r.R.ReadUint32()
	if err != nil {
		return nil, err
	}
	payload, err := r.R.ReadPayload()
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
	sn, err := r.R.ReadUint32()
	if err != nil {
		return nil, err
	}
	key, err := r.R.ReadUint32()
	if err != nil {
		return nil, err
	}
	return &eventing.PerformHandoffKeyMessage{
		Sequence: sn,
		Key:      key,
	}, nil
}

func (r *GatewayMessageReader) ReadPasswordMessage() (*eventing.PerformHandoffPasswordMessage, error) {
	sn, err := r.R.ReadUint32()
	if err != nil {
		return nil, err
	}
	password, err := r.R.ReadString()
	if err != nil {
		return nil, err
	}
	return &eventing.PerformHandoffPasswordMessage{
		Sequence: sn,
		Password: password,
	}, nil
}
