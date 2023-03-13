package protonostale

import (
	"bytes"
	"strconv"

	eventing "github.com/infinity-blackhole/elkia/pkg/api/eventing/v1alpha1"
)

var (
	HeartBeatOpCode = "0"
)

type ChannelInteractRequest struct {
	eventing.ChannelInteractRequest
}

type ChannelFrame struct {
	eventing.ChannelFrame
}

func (e *ChannelFrame) UnmarshalNosTale(b []byte) error {
	fields := bytes.SplitN(b, []byte(" "), 3)
	sn, err := strconv.ParseUint(string(fields[1]), 10, 32)
	if err != nil {
		return err
	}
	e.Sequence = uint32(sn)
	switch string(fields[0]) {
	case HeartBeatOpCode:
		var heartbeat HeartbeatFrame
		if err := heartbeat.UnmarshalNosTale(fields[1]); err != nil {
			return err
		}
		e.Payload = &eventing.ChannelFrame_HeartbeatFrame{
			HeartbeatFrame: &heartbeat.HeartbeatFrame,
		}
	default:
		e.Payload = &eventing.ChannelFrame_RawFrame{
			RawFrame: fields[1],
		}
	}
	return nil
}

type HeartbeatFrame struct {
	eventing.HeartbeatFrame
}

func (e *HeartbeatFrame) MarshalNosTale() ([]byte, error) {
	return []byte{}, nil
}

func (e *HeartbeatFrame) UnmarshalNosTale(b []byte) error {
	return nil
}
