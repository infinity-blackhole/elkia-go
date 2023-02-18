package protonostale

import (
	"bytes"
	"errors"
	"strconv"

	eventing "github.com/infinity-blackhole/elkia/pkg/api/eventing/v1alpha1"
)

func ParseSyncMessage(msg []byte) (*eventing.SyncMessage, error) {
	ss := bytes.Split(msg, []byte(" "))
	if len(ss) != 1 {
		return nil, errors.New("invalid auth message")
	}
	sn, err := ParseUint32(ss[0])
	if err != nil {
		return nil, err
	}
	return &eventing.SyncMessage{
		Sequence: sn,
	}, nil
}

func ParseKeyMessage(
	msg []byte,
) (*eventing.KeyMessage, error) {
	ss := bytes.Split(msg, []byte(" "))
	if len(ss) != 2 {
		return nil, errors.New("invalid auth message")
	}
	sn, err := ParseUint32(ss[0])
	if err != nil {
		return nil, err
	}
	key, err := ParseUint32(ss[1])
	if err != nil {
		return nil, err
	}
	return &eventing.KeyMessage{
		Sequence: sn,
		Key:      key,
	}, nil
}

func ParsePasswordMessage(
	msg []byte,
) (*eventing.PasswordMessage, error) {
	ss := bytes.Split(msg, []byte(" "))
	if len(ss) != 2 {
		return nil, errors.New("invalid auth message")
	}
	sn, err := ParseUint32(ss[0])
	if err != nil {
		return nil, err
	}
	return &eventing.PasswordMessage{
		Sequence: sn,
		Password: string(ss[1]),
	}, nil
}

func ParseUint32(s []byte) (uint32, error) {
	key, err := strconv.ParseUint(string(s), 10, 32)
	if err != nil {
		return 0, err
	}
	return uint32(key), nil
}
