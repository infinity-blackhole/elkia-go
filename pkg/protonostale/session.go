package protonostale

import (
	"bytes"
	"fmt"
	"strconv"

	eventing "github.com/infinity-blackhole/elkia/pkg/api/eventing/v1alpha1"
)

var (
	AuthLoginOpCode = "NoS0575"
)

type AuthInteractRequest struct {
	eventing.AuthInteractRequest
}

func (r *AuthInteractRequest) UnmarshalNosTale(b []byte) error {
	bs := bytes.SplitN(b, []byte(" "), 2)
	opcode := string(bs[0])
	switch opcode {
	case AuthLoginOpCode:
		var authLoginFrame AuthLoginFrame
		if err := authLoginFrame.UnmarshalNosTale(bs[1]); err != nil {
			return err
		}
		r.Payload = &eventing.AuthInteractRequest_LoginFrame{
			LoginFrame: &authLoginFrame.AuthLoginFrame,
		}
	default:
		return fmt.Errorf("invalid opcode: %s", opcode)
	}
	return nil
}

type AuthInteractResponse struct {
	eventing.AuthInteractResponse
}

func (r *AuthInteractResponse) MarshalNosTale() ([]byte, error) {
	var b bytes.Buffer
	switch v := r.Payload.(type) {
	case *eventing.AuthInteractResponse_ErrorFrame:
		if _, err := fmt.Fprintf(&b, "%s ", ErrorOpCode); err != nil {
			return nil, err
		}
		vv := &ErrorFrame{
			ErrorFrame: eventing.ErrorFrame{
				Code: v.ErrorFrame.Code,
			},
		}
		bs, err := vv.MarshalNosTale()
		if err != nil {
			return nil, err
		}
		if _, err := b.Write(bs); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("invalid payload: %v", v)
	}
	return b.Bytes(), nil
}

type AuthLoginFrame struct {
	eventing.AuthLoginFrame
}

func (e *AuthLoginFrame) UnmarshalNosTale(b []byte) error {
	bs := bytes.Split(b, []byte(" "))
	if len(bs) != 4 {
		return fmt.Errorf("invalid length: %d", len(bs))
	}
	e.Identifier = string(bs[1])
	password, err := DecodePassword(bs[2])
	if err != nil {
		return err
	}
	e.Password = password
	clientVersion, err := DecodeClientVersion(bs[3])
	if err != nil {
		return err
	}
	e.ClientVersion = clientVersion
	return nil
}

func DecodePassword(b []byte) (string, error) {
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

func DecodeClientVersion(b []byte) (string, error) {
	fields := bytes.SplitN(b, []byte{'\v'}, 2)
	fields = bytes.Split(fields[1], []byte("."))
	if len(fields) != 4 {
		return "", fmt.Errorf("invalid version: %s", string(b))
	}
	major, err := strconv.ParseUint(string(fields[0]), 10, 32)
	if err != nil {
		return "", err
	}
	minor, err := strconv.ParseUint(string(fields[1]), 10, 32)
	if err != nil {
		return "", err
	}
	patch, err := strconv.ParseUint(string(fields[2]), 10, 32)
	if err != nil {
		return "", err
	}
	build, err := strconv.ParseUint(string(fields[3][:len(fields[3])-1]), 10, 32)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%d.%d.%d+%d", major, minor, patch, build), nil
}

type EndpointListFrame struct {
	eventing.EndpointListFrame
}

func (e *EndpointListFrame) MarshalNosTale() ([]byte, error) {
	var buff bytes.Buffer
	if _, err := fmt.Fprintf(&buff, "NsTeST %d ", e.Code); err != nil {
		return nil, err
	}
	for _, m := range e.Endpoints {
		b, err := (&Endpoint{
			Endpoint: eventing.Endpoint{
				Host:      m.Host,
				Port:      m.Port,
				Weight:    m.Weight,
				WorldId:   m.WorldId,
				ChannelId: m.ChannelId,
				WorldName: m.WorldName,
			},
		}).MarshalNosTale()
		if err != nil {
			return nil, err
		}
		if _, err := buff.Write(b); err != nil {
			return nil, err
		}
	}
	if _, err := buff.WriteString("-1:-1:-1:10000.10000.1\n"); err != nil {
		return nil, err
	}
	return buff.Bytes(), nil
}

func (e *EndpointListFrame) UnmarshalNosTale(b []byte) error {
	bs := bytes.Split(b, []byte(" "))
	if len(bs) < 2 {
		return fmt.Errorf("invalid length: %d", len(bs))
	}
	if string(bs[0]) != "NsTeST" {
		return fmt.Errorf("invalid prefix: %s", string(bs[0]))
	}
	code, err := strconv.ParseUint(string(bs[1]), 10, 32)
	if err != nil {
		return err
	}
	e.Code = uint32(code)
	for _, m := range bs[2:] {
		var ep Endpoint
		if err := ep.UnmarshalNosTale(m); err != nil {
			return err
		}
		e.Endpoints = append(e.Endpoints, &ep.Endpoint)
	}
	return nil
}

type Endpoint struct {
	eventing.Endpoint
}

func (e *Endpoint) MarshalNosTale() ([]byte, error) {
	var buff bytes.Buffer
	if _, err := fmt.Fprintf(
		&buff,
		"%s:%s:%d:%d.%d.%s ",
		e.Host,
		e.Port,
		e.Weight,
		e.WorldId,
		e.ChannelId,
		e.WorldName,
	); err != nil {
		return nil, err
	}
	return buff.Bytes(), nil
}

func (e *Endpoint) UnmarshalNosTale(b []byte) error {
	bs := bytes.Split(b, []byte(":"))
	if len(bs) != 5 {
		return fmt.Errorf("invalid length: %d", len(bs))
	}
	e.Host = string(bs[0])
	e.Port = string(bs[1])
	weight, err := strconv.ParseUint(string(bs[2]), 10, 32)
	if err != nil {
		return err
	}
	e.Weight = uint32(weight)
	bs = bytes.Split(bs[3], []byte("."))
	if len(bs) != 3 {
		return fmt.Errorf("invalid length: %d", len(bs))
	}
	worldId, err := strconv.ParseUint(string(bs[0]), 10, 32)
	if err != nil {
		return err
	}
	e.WorldId = uint32(worldId)
	channelId, err := strconv.ParseUint(string(bs[1]), 10, 32)
	if err != nil {
		return err
	}
	e.ChannelId = uint32(channelId)
	e.WorldName = string(bs[2])
	return nil
}

type SyncFrame struct {
	eventing.SyncFrame
}

func (e *SyncFrame) UnmarshalNosTale(b []byte) error {
	bs := bytes.Split(b[2:], []byte(" "))
	if len(bs) != 6 {
		return fmt.Errorf("invalid length: %d", len(bs))
	}
	sn, err := strconv.ParseUint(string(bs[0]), 10, 32)
	if err != nil {
		return err
	}
	e.Sequence = uint32(sn)
	code, err := strconv.ParseUint(string(bs[1]), 10, 32)
	if err != nil {
		return err
	}
	e.Code = uint32(code)
	return nil
}

type HandoffFrame struct {
	eventing.HandoffFrame
}

func (e *HandoffFrame) MarshalNosTale() ([]byte, error) {
	var b bytes.Buffer
	bs, err := (&HandoffIdentifierFrame{
		HandoffIdentifierFrame: eventing.HandoffIdentifierFrame{
			Sequence:   e.HandoffFrame.IdentifierFrame.Sequence,
			Identifier: e.HandoffFrame.IdentifierFrame.Identifier,
		},
	}).
		MarshalNosTale()
	if err != nil {
		return nil, err
	}
	if _, err := b.Write(bs); err != nil {
		return nil, err
	}
	bs, err = (&HandoffPasswordFrame{
		HandoffPasswordFrame: eventing.HandoffPasswordFrame{
			Sequence: e.PasswordFrame.Sequence,
			Password: e.PasswordFrame.Password,
		},
	}).
		MarshalNosTale()
	if err != nil {
		return nil, err
	}
	if _, err := b.Write(bs); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func (e *HandoffFrame) UnmarshalNosTale(b []byte) error {
	bs := bytes.Split(b, []byte(" "))
	if len(bs) != 5 {
		return fmt.Errorf("invalid length: %d", len(bs))
	}
	var identifier HandoffIdentifierFrame
	if err := identifier.UnmarshalNosTale(bytes.Join(bs[0:2], []byte{' '})); err != nil {
		return err
	}
	e.HandoffFrame.IdentifierFrame = &eventing.HandoffIdentifierFrame{
		Sequence:   identifier.Sequence,
		Identifier: identifier.Identifier,
	}
	var password HandoffPasswordFrame
	if err := password.UnmarshalNosTale(bytes.Join(bs[2:4], []byte{' '})); err != nil {
		return err
	}
	e.HandoffFrame.PasswordFrame = &eventing.HandoffPasswordFrame{
		Sequence: password.Sequence,
		Password: password.Password,
	}
	return nil
}

type HandoffIdentifierFrame struct {
	eventing.HandoffIdentifierFrame
}

func (e *HandoffIdentifierFrame) MarshalNosTale() ([]byte, error) {
	var b bytes.Buffer
	if _, err := fmt.Fprintf(&b, "%d ", e.Sequence); err != nil {
		return nil, err
	}
	if _, err := b.WriteString(e.Identifier); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func (e *HandoffIdentifierFrame) UnmarshalNosTale(b []byte) error {
	bs := bytes.Split(b, []byte(" "))
	if len(bs) != 2 {
		return fmt.Errorf("invalid length: %d", len(bs))
	}
	sn, err := strconv.ParseUint(string(bs[0]), 10, 32)
	if err != nil {
		return err
	}
	e.Sequence = uint32(sn)
	e.Identifier = string(bs[1])
	return nil
}

type HandoffPasswordFrame struct {
	eventing.HandoffPasswordFrame
}

func (e *HandoffPasswordFrame) MarshalNosTale() ([]byte, error) {
	var b bytes.Buffer
	if _, err := fmt.Fprintf(&b, "%d ", e.Sequence); err != nil {
		return nil, err
	}
	if _, err := b.WriteString(e.Password); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func (e *HandoffPasswordFrame) UnmarshalNosTale(b []byte) error {
	bs := bytes.Split(b, []byte(" "))
	if len(bs) != 2 {
		return fmt.Errorf("invalid length: %d", len(bs))
	}
	sn, err := strconv.ParseUint(string(bs[0]), 10, 32)
	if err != nil {
		return err
	}
	e.Sequence = uint32(sn)
	e.Password = string(bs[1])
	return nil
}