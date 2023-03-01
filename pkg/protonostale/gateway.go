package protonostale

import (
	"bytes"

	eventing "github.com/infinity-blackhole/elkia/pkg/api/eventing/v1alpha1"
)

func DecodeAuthHandoffSyncEvent(
	b []byte,
) (*eventing.AuthHandoffSyncEvent, error) {
	fields := bytes.Fields(b[2:])
	sn, err := DecodeUint(fields[0])
	if err != nil {
		return nil, err
	}
	code, err := DecodeUint(fields[1])
	if err != nil {
		return nil, err
	}
	return &eventing.AuthHandoffSyncEvent{
		Sequence: sn,
		Code:     code,
	}, nil
}

func DecodeAuthHandoffLoginEvent(
	b []byte,
) (*eventing.AuthHandoffLoginEvent, error) {
	fields := bytes.Fields(b)
	idMsg, err := DecodeAuthHandoffLoginIdentifierEvent(fields[0])
	if err != nil {
		return nil, err
	}
	pwdMsg, err := DecodeAuthHandoffLoginPasswordEvent(fields[1])
	if err != nil {
		return nil, err
	}
	return &eventing.AuthHandoffLoginEvent{
		IdentifierEvent: idMsg,
		PasswordEvent:   pwdMsg,
	}, nil
}

func DecodeAuthHandoffLoginIdentifierEvent(
	b []byte,
) (*eventing.AuthHandoffLoginIdentifierEvent, error) {
	fields := bytes.Fields(b)
	sn, err := DecodeUint(fields[0])
	if err != nil {
		return nil, err
	}
	return &eventing.AuthHandoffLoginIdentifierEvent{
		Sequence:   sn,
		Identifier: string(fields[1]),
	}, nil
}

func DecodeAuthHandoffLoginPasswordEvent(
	b []byte,
) (*eventing.AuthHandoffLoginPasswordEvent, error) {
	fields := bytes.Fields(b)
	sn, err := DecodeUint(fields[0])
	if err != nil {
		return nil, err
	}
	return &eventing.AuthHandoffLoginPasswordEvent{
		Sequence: sn,
		Password: string(fields[1]),
	}, nil
}

func DecodeChannelEvent(b []byte) (*eventing.ChannelEvent, error) {
	fields := bytes.Fields(b)
	sn, err := DecodeUint(fields[0])
	if err != nil {
		return nil, err
	}
	return &eventing.ChannelEvent{
		Sequence: sn,
		Payload:  &eventing.ChannelEvent_UnknownPayload{UnknownPayload: fields[1]},
	}, nil
}
