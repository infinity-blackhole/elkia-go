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

func ParsePerformHandoffMessage(
	msg []byte,
) (*eventing.PerformHandoffMessage, error) {
	ss := bytes.Split(msg, []byte(" "))
	if len(ss) != 4 {
		return nil, errors.New("invalid auth message")
	}
	sni, err := ParseUint32(ss[0])
	if err != nil {
		return nil, err
	}
	key, err := ParseUint32(ss[1])
	if err != nil {
		return nil, err
	}
	snp, err := ParseUint32(ss[2])
	if err != nil {
		return nil, err
	}
	return &eventing.PerformHandoffMessage{
		KeySequence:      snp,
		Key:              key,
		PasswordSequence: sni,
		Password:         string(ss[3]),
	}, nil
}

func ParseUint32(s []byte) (uint32, error) {
	key, err := strconv.ParseUint(string(s), 10, 32)
	if err != nil {
		return 0, err
	}
	return uint32(key), nil
}
