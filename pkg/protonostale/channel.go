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

type ChannelEvent struct {
	*eventing.ChannelEvent
}

type ClientInteractRequest struct {
	*eventing.ClientInteractRequest
}

func (f *ClientInteractRequest) UnmarshalNosTale(b []byte) error {
	f.ClientInteractRequest = &eventing.ClientInteractRequest{}
	fields := bytes.SplitN(b, FieldSeparator, 3)
	sn, err := strconv.ParseUint(string(fields[0]), 10, 32)
	if err != nil {
		return err
	}
	f.Sequence = uint32(sn)
	switch {
	case bytes.Equal(HeartbeatOpCode, fields[1]):
		f.Command = &eventing.ClientCommand{
			Command: &eventing.ClientCommand_Heartbeat{},
		}
	}
	return nil
}
