package protonostale

import (
	"bufio"

	eventing "github.com/infinity-blackhole/elkia/pkg/api/eventing/v1alpha1"
)

func ReadSyncEvent(r *bufio.Reader) (*eventing.SyncEvent, error) {
	sn, err := ReadSequence(r)
	if err != nil {
		return nil, err
	}
	return &eventing.SyncEvent{Sequence: sn}, nil
}

func ReadAuthHandoffEvent(r *bufio.Reader) (*eventing.AuthHandoffEvent, error) {
	keyMsg, err := ReadAuthHandoffKeyEvent(r)
	if err != nil {
		return nil, err
	}
	pwdMsg, err := ReadAuthHandoffPasswordEvent(r)
	if err != nil {
		return nil, err
	}
	return &eventing.AuthHandoffEvent{
		KeyEvent:      keyMsg,
		PasswordEvent: pwdMsg,
	}, nil
}

func ReadAuthHandoffKeyEvent(r *bufio.Reader) (*eventing.AuthHandoffKeyEvent, error) {
	sn, err := ReadSequence(r)
	if err != nil {
		return nil, err
	}
	key, err := ReadUint32(r)
	if err != nil {
		return nil, err
	}
	return &eventing.AuthHandoffKeyEvent{
		Sequence: sn,
		Key:      key,
	}, nil
}

func ReadAuthHandoffPasswordEvent(r *bufio.Reader) (*eventing.AuthHandoffPasswordEvent, error) {
	sn, err := ReadSequence(r)
	if err != nil {
		return nil, err
	}
	password, err := ReadString(r)
	if err != nil {
		return nil, err
	}
	return &eventing.AuthHandoffPasswordEvent{
		Sequence: sn,
		Password: password,
	}, nil
}

func ReadChannelEvent(r *bufio.Reader) (*eventing.ChannelEvent, error) {
	sn, err := ReadSequence(r)
	if err != nil {
		return nil, err
	}
	opcode, err := ReadOpCode(r)
	if err != nil {
		return nil, err
	}
	payload, err := ReadPayload(r)
	if err != nil {
		return nil, err
	}
	return &eventing.ChannelEvent{
		Sequence: sn,
		OpCode:   opcode,
		Payload:  &eventing.ChannelEvent_UnknownPayload{UnknownPayload: payload},
	}, nil
}
