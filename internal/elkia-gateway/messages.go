package elkiagateway

import (
	"errors"
	"strconv"
	"strings"
)

type Message struct {
	SequenceNumber uint32
}

type SyncMessage struct {
	*Message
}

func ParseSyncMessage(msg string) (*SyncMessage, error) {
	ss := strings.Split(msg, " ")
	if len(ss) != 1 {
		return nil, errors.New("invalid auth message")
	}
	sn, err := ParseUint32(ss[0])
	if err != nil {
		return nil, err
	}
	return &SyncMessage{
		Message: &Message{
			SequenceNumber: sn,
		},
	}, nil
}

type CredentialsMessage struct {
	*Message
	Key      uint32
	Password string
}

func ParseCredentialsMessage(msg string) (*CredentialsMessage, error) {
	ss := strings.Split(msg, " ")
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
	if sni != snp+1 {
		return nil, errors.New("sequence number mismatch")
	}
	return &CredentialsMessage{
		Message:  &Message{SequenceNumber: snp},
		Key:      key,
		Password: ss[3],
	}, nil
}

func ParseUint32(s string) (uint32, error) {
	key, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint32(key), nil
}
