package protonostale

import (
	"bytes"
	"strconv"

	eventing "github.com/infinity-blackhole/elkia/pkg/api/eventing/v1alpha1"
)

var (
	SyncOpCode       = "sync"
	IdentifierOpCode = "id"
	PasswordOpCode   = "pass"
	HeartBeatOpCode  = "0"
)

type ChannelInteractRequest struct {
	*eventing.ChannelInteractRequest
}

func (e *ChannelInteractRequest) MarshalNosTale() ([]byte, error) {
	return nil, nil
}

func (e *ChannelInteractRequest) UnmarshalNosTale(b []byte) error {
	fields := bytes.SplitN(b, []byte(" "), 3)
	sn, err := strconv.ParseUint(string(fields[0]), 10, 32)
	if err != nil {
		return err
	}
	e.Sequence = uint32(sn)
	switch string(fields[1]) {
	case HeartBeatOpCode:
		var heartbeat HeartbeatFrame
		if err := heartbeat.UnmarshalNosTale(fields[1]); err != nil {
			return err
		}
		e.Payload = &eventing.ChannelInteractRequest_HeartbeatFrame{
			HeartbeatFrame: heartbeat.HeartbeatFrame,
		}
	case SyncOpCode:
		var identifier SyncFrame
		if err := identifier.UnmarshalNosTale(fields[1]); err != nil {
			return err
		}
		e.Payload = &eventing.ChannelInteractRequest_SyncFrame{
			SyncFrame: identifier.SyncFrame,
		}
	case IdentifierOpCode:
		var identifier IdentifierFrame
		if err := identifier.UnmarshalNosTale(fields[1]); err != nil {
			return err
		}
		e.Payload = &eventing.ChannelInteractRequest_IdentifierFrame{
			IdentifierFrame: identifier.IdentifierFrame,
		}
	case PasswordOpCode:
		var identifier PasswordFrame
		if err := identifier.UnmarshalNosTale(fields[1]); err != nil {
			return err
		}
		e.Payload = &eventing.ChannelInteractRequest_PasswordFrame{
			PasswordFrame: identifier.PasswordFrame,
		}
	default:
		e.Payload = &eventing.ChannelInteractRequest_RawFrame{
			RawFrame: fields[1],
		}
	}
	return nil
}

type HeartbeatFrame struct {
	*eventing.HeartbeatFrame
}

func (e *HeartbeatFrame) MarshalNosTale() ([]byte, error) {
	return []byte{}, nil
}

func (e *HeartbeatFrame) UnmarshalNosTale(b []byte) error {
	return nil
}
