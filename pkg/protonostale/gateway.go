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
	sn, err := ParseUint(s)
	if err != nil {
		return nil, err
	}
	return &eventing.AuthHandoffSyncEvent{Sequence: sn}, nil
}

func ParseAuthHandoffLoginEvent(
	r string,
) (*eventing.AuthHandoffLoginEvent, error) {
	fields := strings.Fields(r)
	keyMsg, err := ParseAuthHandoffLoginKeyEvent(fields[0])
	if err != nil {
		return nil, err
	}
	pwdMsg, err := ParseAuthHandoffLoginPasswordEvent(fields[1])
	if err != nil {
		return nil, err
	}
	return &eventing.AuthHandoffLoginEvent{
		KeyEvent:      keyMsg,
		PasswordEvent: pwdMsg,
	}, nil
}

func ParseAuthHandoffLoginKeyEvent(
	s string,
) (*eventing.AuthHandoffLoginKeyEvent, error) {
	fields := strings.Fields(s)
	sn, err := ParseUint(fields[0])
	if err != nil {
		return nil, err
	}
	key, err := ParseUint(fields[1])
	if err != nil {
		return nil, err
	}
	return &eventing.AuthHandoffLoginKeyEvent{
		Sequence: sn,
		Key:      key,
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
		return nil, fmt.Errorf("failed to read channel event: %w", s.Err())
	}
	sn, err := ParseUint(s.Text())
	if err != nil {
		return nil, err
	}
	if !s.Scan() {
		return nil, fmt.Errorf("failed to read channel event: %w", s.Err())
	}
	_ = s.Text()
	if !s.Scan() {
		return nil, fmt.Errorf("failed to read channel event: %w", s.Err())
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
