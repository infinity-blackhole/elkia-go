package protonostale

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"

	eventing "github.com/infinity-blackhole/elkia/pkg/api/eventing/v1alpha1"
)

func ParseAuthHandoffSyncEvent(
	s string,
) (*eventing.AuthHandoffSyncEvent, error) {
	fields := strings.Fields(s[2:])
	sn, err := ParseUint(fields[0])
	if err != nil {
		return nil, err
	}
	key, err := ParseUint(fields[1])
	if err != nil {
		return nil, err
	}
	return &eventing.AuthHandoffSyncEvent{
		Sequence: sn,
		Key:      key,
	}, nil
}

func ParseAuthHandoffLoginEvent(
	r string,
) (*eventing.AuthHandoffLoginEvent, error) {
	fields := strings.Fields(r)
	idMsg, err := ParseAuthHandoffLoginIdentifierEvent(fields[0])
	if err != nil {
		return nil, err
	}
	pwdMsg, err := ParseAuthHandoffLoginPasswordEvent(fields[1])
	if err != nil {
		return nil, err
	}
	return &eventing.AuthHandoffLoginEvent{
		IdentifierEvent: idMsg,
		PasswordEvent:   pwdMsg,
	}, nil
}

func ParseAuthHandoffLoginIdentifierEvent(
	s string,
) (*eventing.AuthHandoffLoginIdentifierEvent, error) {
	fields := strings.Fields(s)
	sn, err := ParseUint(fields[0])
	if err != nil {
		return nil, err
	}
	return &eventing.AuthHandoffLoginIdentifierEvent{
		Sequence:   sn,
		Identifier: fields[1],
	}, nil
}

func ParseAuthHandoffLoginPasswordEvent(
	s string,
) (*eventing.AuthHandoffLoginPasswordEvent, error) {
	fields := strings.Fields(s)
	sn, err := ParseUint(fields[0])
	if err != nil {
		return nil, err
	}
	return &eventing.AuthHandoffLoginPasswordEvent{
		Sequence: sn,
		Password: fields[1],
	}, nil
}

func ReadChannelEvent(r *bufio.Reader) (*eventing.ChannelEvent, error) {
	s := bufio.NewScanner(r)
	s.Split(bufio.ScanWords)
	if !s.Scan() {
		if err := s.Err(); err != nil {
			return nil, fmt.Errorf("failed to read channel event: %w", err)
		}
		return nil, fmt.Errorf("failed to read channel event: EOF")
	}
	sn, err := ParseUint(s.Text())
	if err != nil {
		return nil, err
	}
	if !s.Scan() {
		if err := s.Err(); err != nil {
			return nil, fmt.Errorf("failed to read channel event: %w", err)
		}
		return nil, fmt.Errorf("failed to read channel event: EOF")
	}
	_ = s.Text()
	if !s.Scan() {
		if err := s.Err(); err != nil {
			return nil, fmt.Errorf("failed to read channel event: %w", err)
		}
		return nil, fmt.Errorf("failed to read channel event: EOF")
	}
	payload := s.Bytes()
	return &eventing.ChannelEvent{
		Sequence: sn,
		Payload:  &eventing.ChannelEvent_UnknownPayload{UnknownPayload: payload},
	}, nil
}

func ScanPackedEvents(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.IndexByte(data, 0xFF); i >= 0 {
		return i + 1, data[0:i], nil
	}
	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), data, nil
	}
	// Request more data.
	return 0, nil, nil
}
