package protonostale

import (
	"bytes"

	eventing "github.com/infinity-blackhole/elkia/pkg/api/eventing/v1alpha1"
)

func DecodeAuthHandoffSyncFrame(
	b []byte,
) (*eventing.AuthHandoffSyncFrame, error) {
	fields := bytes.Fields(b[2:])
	sn, err := DecodeUint(fields[0])
	if err != nil {
		return nil, err
	}
	code, err := DecodeUint(fields[1])
	if err != nil {
		return nil, err
	}
	return &eventing.AuthHandoffSyncFrame{
		Sequence: sn,
		Code:     code,
	}, nil
}

func DecodeAuthHandoffLoginFrame(
	b []byte,
) (*eventing.AuthHandoffLoginFrame, error) {
	fields := bytes.Fields(b)
	idMsg, err := DecodeAuthHandoffLoginIdentifierFrame(fields[0])
	if err != nil {
		return nil, err
	}
	pwdMsg, err := DecodeAuthHandoffLoginPasswordFrame(fields[1])
	if err != nil {
		return nil, err
	}
	return &eventing.AuthHandoffLoginFrame{
		IdentifierFrame: idMsg,
		PasswordFrame:   pwdMsg,
	}, nil
}

func DecodeAuthHandoffLoginIdentifierFrame(
	b []byte,
) (*eventing.AuthHandoffLoginIdentifierFrame, error) {
	fields := bytes.Fields(b)
	sn, err := DecodeUint(fields[0])
	if err != nil {
		return nil, err
	}
	return &eventing.AuthHandoffLoginIdentifierFrame{
		Sequence:   sn,
		Identifier: string(fields[1]),
	}, nil
}

func DecodeAuthHandoffLoginPasswordFrame(
	b []byte,
) (*eventing.AuthHandoffLoginPasswordFrame, error) {
	fields := bytes.Fields(b)
	sn, err := DecodeUint(fields[0])
	if err != nil {
		return nil, err
	}
	return &eventing.AuthHandoffLoginPasswordFrame{
		Sequence: sn,
		Password: string(fields[1]),
	}, nil
}

func DecodeChannelFrame(b []byte) (*eventing.ChannelFrame, error) {
	fields := bytes.Fields(b)
	sn, err := DecodeUint(fields[0])
	if err != nil {
		return nil, err
	}
	return &eventing.ChannelFrame{
		Sequence: sn,
		Payload:  &eventing.ChannelFrame_UnknownPayload{UnknownPayload: fields[1]},
	}, nil
}
