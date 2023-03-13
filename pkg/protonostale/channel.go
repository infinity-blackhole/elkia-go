package protonostale

import (
	"bytes"
	"fmt"
	"strconv"

	eventing "github.com/infinity-blackhole/elkia/pkg/api/eventing/v1alpha1"
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
	var b bytes.Buffer
	if _, err := fmt.Fprintf(&b, "%d ", e.Sequence); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func (e *HeartbeatFrame) UnmarshalNosTale(b []byte) error {
	fields := bytes.Split(b, []byte(" "))
	if len(fields) != 2 {
		return fmt.Errorf("invalid length: %d", len(fields))
	}
	sn, err := strconv.ParseUint(string(fields[0]), 10, 32)
	if err != nil {
		return err
	}
	e.Sequence = uint32(sn)
	return nil
}
