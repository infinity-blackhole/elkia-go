package protonostale

import eventing "go.shikanime.studio/elkia/pkg/api/eventing/v1alpha1"

type HeartbeatCommand struct {
	*eventing.HeartbeatCommand
}

func (f *HeartbeatCommand) UnmarshalNosTale(b []byte) error {
	return nil
}
