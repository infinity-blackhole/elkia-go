package protonostale

import (
	"bytes"
	"strconv"

	eventing "github.com/infinity-blackhole/elkia/pkg/api/eventing/v1alpha1"
)

var (
	HeartbeatOpCode = []byte("0")
)

type ChannelInteractRequest struct {
	*eventing.ChannelInteractRequest
}

type ChannelInteractResponse struct {
	*eventing.ChannelInteractResponse
}

type CommandFrame struct {
	*eventing.CommandFrame
}

func (f *CommandFrame) UnmarshalNosTale(b []byte) error {
	f.CommandFrame = &eventing.CommandFrame{}
	fields := bytes.SplitN(b, FieldSeparator, 3)
	sn, err := strconv.ParseUint(string(fields[0]), 10, 32)
	if err != nil {
		return err
	}
	f.Sequence = uint32(sn)
	switch {
	case bytes.Equal(HeartbeatOpCode, fields[1]):
		f.Payload = &eventing.CommandFrame_HeartbeatFrame{
			HeartbeatFrame: &eventing.HeartbeatFrame{},
		}
	default:
		f.Payload = &eventing.CommandFrame_RawFrame{
			RawFrame: fields[1],
		}
	}
	return nil
}

type HeartbeatFrame struct {
	*eventing.HeartbeatFrame
}

func (f *HeartbeatFrame) MarshalNosTale() ([]byte, error) {
	return []byte{}, nil
}

func (f *HeartbeatFrame) UnmarshalNosTale(b []byte) error {
	return nil
}
