package protonostale

import (
	"bytes"
	"fmt"
	"strconv"

	eventing "github.com/infinity-blackhole/elkia/pkg/api/eventing/v1alpha1"
)

var (
	HeartBeatOpCode = "0"
)

type ChannelInteractRequest struct {
	eventing.ChannelInteractRequest
}

func (e *ChannelInteractRequest) MarshalNosTale() ([]byte, error) {
	return nil, nil
}

func (e *ChannelInteractRequest) UnmarshalNosTale(b []byte) error {
	bs := bytes.SplitN(b, []byte(" "), 2)
	opcode := string(bs[0])
	switch opcode {
	case HeartBeatOpCode:
		var heartbeat HeartbeatFrame
		if err := heartbeat.UnmarshalNosTale(bs[1]); err != nil {
			return err
		}
		e.Payload = &eventing.ChannelInteractRequest_HeartbeatFrame{}
	default:
		e.Payload = &eventing.ChannelInteractRequest_RawFrame{
			RawFrame: &eventing.RawFrame{
				Sequence: 0,
				Payload:  bs[1],
			},
		}
	}
	return nil
}

type HeartbeatFrame struct {
	eventing.HeartbeatFrame
}

func (e *HeartbeatFrame) MarshalNosTale() ([]byte, error) {
	var b bytes.Buffer
	if _, err := fmt.Fprintf(&b, "%d ", e.Sequence); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func (e *HeartbeatFrame) UnmarshalNosTale(b []byte) error {
	bs := bytes.SplitN(b, []byte(" "), 2)
	sn, err := strconv.ParseUint(string(bs[0]), 10, 32)
	if err != nil {
		return err
	}
	e.Sequence = uint32(sn)
	return nil
}

type WorldFrame struct {
	eventing.WorldFrame
}

func (e *WorldFrame) MarshalNosTale() ([]byte, error) {
	var b bytes.Buffer
	if _, err := fmt.Fprintf(&b, "%d ", e.Sequence); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func (e *WorldFrame) UnmarshalNosTale(b []byte) error {
	bs := bytes.SplitN(b, []byte(" "), 2)
	sn, err := strconv.ParseUint(string(bs[0]), 10, 32)
	if err != nil {
		return err
	}
	e.Sequence = uint32(sn)
	return nil
}
