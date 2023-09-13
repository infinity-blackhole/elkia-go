package protonostale

import eventingpb "go.shikanime.studio/elkia/pkg/api/eventing/v1alpha1"

type HeartbeatCommand struct {
	*eventingpb.HeartbeatCommand
}

func (f *HeartbeatCommand) UnmarshalNosTale(b []byte) error {
	return nil
}
