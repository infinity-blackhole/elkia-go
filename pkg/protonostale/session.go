package protonostale

import (
	"bytes"
	"fmt"
	"strconv"

	eventingpb "go.shikanime.studio/elkia/pkg/api/eventing/v1alpha1"
)

var (
	CreateHandoffFlowOpCode = "NoS0575"
)

type AuthCommand struct {
	*eventingpb.AuthCommand
}

func (f *AuthCommand) UnmarshalNosTale(b []byte) error {
	f.AuthCommand = &eventingpb.AuthCommand{}
	fields := bytes.SplitN(b, FieldSeparator, 2)
	opcode := string(fields[0])
	switch opcode {
	case CreateHandoffFlowOpCode:
		var CreateHandoffFlowCommand CreateHandoffFlowCommand
		if err := CreateHandoffFlowCommand.UnmarshalNosTale(fields[1]); err != nil {
			return err
		}
		f.Command = &eventingpb.AuthCommand_CreateHandoffFlow{
			CreateHandoffFlow: CreateHandoffFlowCommand.CreateHandoffFlowCommand,
		}
	default:
		return fmt.Errorf("invalid opcode: %s", opcode)
	}
	return nil
}

type ClientEvent struct {
	*eventingpb.ClientEvent
}

func (f *ClientEvent) MarshalNosTale() ([]byte, error) {
	var (
		fields []byte
		err    error
	)
	switch p := f.Event.(type) {
	case *eventingpb.ClientEvent_Error:
		fields, err = MarshalNosTale(&ErrorEvent{
			ErrorEvent: p.Error,
		})
	case *eventingpb.ClientEvent_Info:
		fields, err = MarshalNosTale(&InfoEvent{
			InfoEvent: p.Info,
		})
	default:
		return nil, fmt.Errorf("invalid payload: %v", p)
	}
	if err != nil {
		return nil, err
	}
	var buff bytes.Buffer
	if _, err := buff.Write(fields); err != nil {
		return nil, err
	}
	return buff.Bytes(), nil
}

type AuthEvent struct {
	*eventingpb.AuthEvent
}

func (f *AuthEvent) MarshalNosTale() ([]byte, error) {
	var (
		fields []byte
		err    error
	)
	switch p := f.Event.(type) {
	case *eventingpb.AuthEvent_Client:
		fields, err = MarshalNosTale(&ClientEvent{
			ClientEvent: p.Client,
		})
	case *eventingpb.AuthEvent_Presence:
		fields, err = MarshalNosTale(&PresenceEvent{
			PresenceEvent: p.Presence,
		})
	default:
		return nil, fmt.Errorf("invalid payload: %v", p)
	}
	if err != nil {
		return nil, err
	}
	var buff bytes.Buffer
	if _, err := buff.Write(fields); err != nil {
		return nil, err
	}
	return buff.Bytes(), nil
}

type CreateHandoffFlowCommand struct {
	*eventingpb.CreateHandoffFlowCommand
}

func (f *CreateHandoffFlowCommand) UnmarshalNosTale(b []byte) error {
	f.CreateHandoffFlowCommand = &eventingpb.CreateHandoffFlowCommand{}
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
