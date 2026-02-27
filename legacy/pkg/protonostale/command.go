package protonostale

import (
	"bytes"
	"strconv"

	eventing "go.shikanime.studio/elkia/pkg/api/eventing/v1alpha1"
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

type CommandCommand struct {
	*eventing.CommandCommand
}

func (f *CommandCommand) UnmarshalNosTale(b []byte) error {
	f.CommandCommand = &eventing.CommandCommand{}
	fields := bytes.SplitN(b, FieldSeparator, 3)
	sn, err := strconv.ParseUint(string(fields[0]), 10, 32)
	if err != nil {
		return err
	}
	f.Sequence = uint32(sn)
	switch {
	case bytes.Equal(HeartbeatOpCode, fields[1]):
		f.Payload = &eventing.CommandCommand_HeartbeatCommand{
			HeartbeatCommand: &eventing.HeartbeatCommand{},
		}
	default:
		f.Payload = &eventing.CommandCommand_RawCommand{
			RawCommand: fields[1],
		}
	}
	return nil
}

type HeartbeatCommand struct {
	*eventing.HeartbeatCommand
}

func (f *HeartbeatCommand) MarshalNosTale() ([]byte, error) {
	return []byte{}, nil
}

func (f *HeartbeatCommand) UnmarshalNosTale(b []byte) error {
	return nil
}
