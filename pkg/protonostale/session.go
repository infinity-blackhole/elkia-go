package protonostale

import (
	"bytes"
	"fmt"
	"strconv"

	eventing "go.shikanime.studio/elkia/pkg/api/eventing/v1alpha1"
)

var (
	AuthCreateHandoffFlowOpCode = "NoS0575"
)

type AuthInteractRequest struct {
	*eventing.AuthInteractRequest
}

func (f *AuthInteractRequest) UnmarshalNosTale(b []byte) error {
	f.AuthInteractRequest = &eventing.AuthInteractRequest{}
	fields := bytes.SplitN(b, FieldSeparator, 2)
	opcode := string(fields[0])
	switch opcode {
	case AuthCreateHandoffFlowOpCode:
		var LoginCommand LoginCommand
		if err := LoginCommand.UnmarshalNosTale(fields[1]); err != nil {
			return err
		}
		f.Payload = &eventing.AuthInteractRequest_LoginCommand{
			LoginCommand: LoginCommand.LoginCommand,
		}
	default:
		return fmt.Errorf("invalid opcode: %s", opcode)
	}
	return nil
}

type AuthInteractResponse struct {
	*eventing.AuthInteractResponse
}

func (f *AuthInteractResponse) MarshalNosTale() ([]byte, error) {
	var buff bytes.Buffer
	switch p := f.Payload.(type) {
	case *eventing.AuthInteractResponse_ErrorEvent:
		vv := &ErrorEvent{
			ErrorEvent: p.ErrorEvent,
		}
		fields, err := vv.MarshalNosTale()
		if err != nil {
			return nil, err
		}
		if _, err := buff.Write(fields); err != nil {
			return nil, err
		}
	case *eventing.AuthInteractResponse_InfoEvent:
		vv := &InfoEvent{
			InfoEvent: p.InfoEvent,
		}
		fields, err := vv.MarshalNosTale()
		if err != nil {
			return nil, err
		}
		if _, err := buff.Write(fields); err != nil {
			return nil, err
		}
	case *eventing.AuthInteractResponse_EndpointListEvent:
		vv := &EndpointListEvent{
			EndpointListEvent: p.EndpointListEvent,
		}
		fields, err := vv.MarshalNosTale()
		if err != nil {
			return nil, err
		}
		if _, err := buff.Write(fields); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("invalid payload: %v", p)
	}
	return buff.Bytes(), nil
}

type LoginCommand struct {
	*eventing.LoginCommand
}

func (f *LoginCommand) UnmarshalNosTale(b []byte) error {
	f.LoginCommand = &eventing.LoginCommand{}
	fields := bytes.Split(b, FieldSeparator)
	if len(fields) != 4 {
		return fmt.Errorf("invalid length: %d", len(fields))
	}
	f.Identifier = string(fields[1])
	password, err := DecodePassword(fields[2])
	if err != nil {
		return err
	}
	f.Password = password
	clientVersion, err := DecodeClientVersion(fields[3])
	if err != nil {
		return err
	}
	f.ClientVersion = clientVersion
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
	build, err := strconv.ParseUint(string(fields[3]), 10, 32)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%d.%d.%d+%d", major, minor, patch, build), nil
}

type EndpointListEvent struct {
	*eventing.EndpointListEvent
}

func (f *EndpointListEvent) MarshalNosTale() ([]byte, error) {
	var buff bytes.Buffer
	if _, err := fmt.Fprintf(&buff, "NsTeST %d ", f.Code); err != nil {
		return nil, err
	}
	for _, m := range f.Endpoints {
		endpoint := &Endpoint{
			Endpoint: &eventing.Endpoint{
				Host:      m.Host,
				Port:      m.Port,
				Weight:    m.Weight,
				WorldId:   m.WorldId,
				ChannelId: m.ChannelId,
				WorldName: m.WorldName,
			},
		}
		b, err := endpoint.MarshalNosTale()
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

func (f *EndpointListEvent) UnmarshalNosTale(b []byte) error {
	f.EndpointListEvent = &eventing.EndpointListEvent{}
	fields := bytes.Split(b, FieldSeparator)
	if len(fields) < 2 {
		return fmt.Errorf("invalid length: %d", len(fields))
	}
	if string(fields[0]) != "NsTeST" {
		return fmt.Errorf("invalid prefix: %s", string(fields[0]))
	}
	code, err := strconv.ParseUint(string(fields[1]), 10, 32)
	if err != nil {
		return err
	}
	f.Code = uint32(code)
	for _, m := range fields[2:] {
		var ep Endpoint
		if err := ep.UnmarshalNosTale(m); err != nil {
			return err
		}
		f.Endpoints = append(f.Endpoints, ep.Endpoint)
	}
	return nil
}

type Endpoint struct {
	*eventing.Endpoint
}

func (f *Endpoint) MarshalNosTale() ([]byte, error) {
	var buff bytes.Buffer
	if _, err := fmt.Fprintf(
		&buff,
		"%s:%s:%d:%d.%d.%s ",
		f.Host,
		f.Port,
		f.Weight,
		f.WorldId,
		f.ChannelId,
		f.WorldName,
	); err != nil {
		return nil, err
	}
	return buff.Bytes(), nil
}

func (f *Endpoint) UnmarshalNosTale(b []byte) error {
	f.Endpoint = &eventing.Endpoint{}
	fields := bytes.Split(b, []byte(":"))
	if len(fields) != 5 {
		return fmt.Errorf("invalid length: %d", len(fields))
	}
	f.Host = string(fields[0])
	f.Port = string(fields[1])
	weight, err := strconv.ParseUint(string(fields[2]), 10, 32)
	if err != nil {
		return err
	}
	f.Weight = uint32(weight)
	fields = bytes.Split(fields[3], []byte("."))
	if len(fields) != 3 {
		return fmt.Errorf("invalid length: %d", len(fields))
	}
	worldId, err := strconv.ParseUint(string(fields[0]), 10, 32)
	if err != nil {
		return err
	}
	f.WorldId = uint32(worldId)
	channelId, err := strconv.ParseUint(string(fields[1]), 10, 32)
	if err != nil {
		return err
	}
	f.ChannelId = uint32(channelId)
	f.WorldName = string(fields[2])
	return nil
}

type SyncCommand struct {
	*eventing.SyncCommand
}

func (f *SyncCommand) UnmarshalNosTale(b []byte) error {
	f.SyncCommand = &eventing.SyncCommand{}
	fields := bytes.Split(b, FieldSeparator)
	if len(fields) != 3 {
		return fmt.Errorf("invalid length: %d", len(fields))
	}
	sn, err := strconv.ParseUint(string(fields[0][2:]), 10, 32)
	if err != nil {
		return err
	}
	f.Sequence = uint32(sn)
	code, err := strconv.ParseUint(string(fields[1]), 10, 32)
	if err != nil {
		return err
	}
	f.Code = uint32(code)
	return nil
}

type IdentifierCommand struct {
	*eventing.IdentifierCommand
}

func (f *IdentifierCommand) MarshalNosTale() ([]byte, error) {
	var b bytes.Buffer
	if _, err := fmt.Fprintf(&b, "%d ", f.Sequence); err != nil {
		return nil, err
	}
	if _, err := b.WriteString(f.Identifier); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func (f *IdentifierCommand) UnmarshalNosTale(b []byte) error {
	f.IdentifierCommand = &eventing.IdentifierCommand{}
	fields := bytes.Split(b, FieldSeparator)
	if len(fields) != 2 {
		return fmt.Errorf("invalid length: %d", len(fields))
	}
	sn, err := strconv.ParseUint(string(fields[0]), 10, 32)
	if err != nil {
		return err
	}
	f.Sequence = uint32(sn)
	f.Identifier = string(fields[1])
	return nil
}

type PasswordCommand struct {
	*eventing.PasswordCommand
}

func (f *PasswordCommand) MarshalNosTale() ([]byte, error) {
	var b bytes.Buffer
	if _, err := fmt.Fprintf(&b, "%d ", f.Sequence); err != nil {
		return nil, err
	}
	if _, err := b.WriteString(f.Password); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func (f *PasswordCommand) UnmarshalNosTale(b []byte) error {
	f.PasswordCommand = &eventing.PasswordCommand{}
	fields := bytes.Split(b, FieldSeparator)
	if len(fields) != 2 {
		return fmt.Errorf("invalid length: %d", len(fields))
	}
	sn, err := strconv.ParseUint(string(fields[0]), 10, 32)
	if err != nil {
		return err
	}
	f.Sequence = uint32(sn)
	f.Password = string(fields[1])
	return nil
}
