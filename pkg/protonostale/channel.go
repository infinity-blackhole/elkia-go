package protonostale

import (
	"bytes"
	"strconv"

	eventingpb "go.shikanime.studio/elkia/pkg/api/eventing/v1alpha1"
)

var (
	HeartbeatOpCode = []byte("0")
)

type GatewayCommand struct {
	*eventingpb.GatewayCommand
}

type GatewayEvent struct {
	*eventingpb.GatewayEvent
}

type ClientCommand struct {
	*eventingpb.ClientCommand
}

func (f *ClientCommand) UnmarshalNosTale(b []byte) error {
	f.ClientCommand = &eventingpb.ClientCommand{}
	fields := bytes.SplitN(b, FieldSeparator, 3)
	sn, err := strconv.ParseUint(string(fields[0]), 10, 32)
	if err != nil {
		return err
	}
	f.Sequence = uint32(sn)
	switch {
	case bytes.Equal(HeartbeatOpCode, fields[1]):
		f.Command = &eventingpb.ClientCommand_Heartbeat{}
	}
	return nil
}
