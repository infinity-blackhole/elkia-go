package protonostale

import (
	"bytes"
	"errors"
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
	default:
		return errors.New("protonostale: unknown client interact request")
	}
}

type CoreInteractRequest struct {
	*eventing.CoreInteractRequest
}

func (f *CoreInteractRequest) UnmarshalNosTale(b []byte) error {
	f.CoreInteractRequest = &eventing.CoreInteractRequest{}
	fields := bytes.SplitN(b, FieldSeparator, 3)
	sn, err := strconv.ParseUint(string(fields[0]), 10, 32)
	if err != nil {
		return err
	}
	f.Sequence = uint32(sn)
	switch {
	case bytes.Equal(HeartbeatOpCode, fields[1]):
		f.Request = &eventing.CoreInteractRequest_Heartbeat{
			Heartbeat: &eventing.HeartbeatRequest{},
		}
	default:
		return errors.New("protonostale: unknown client interact request")
	}
	return nil
}

type HeartbeatRequest struct {
	*eventing.HeartbeatRequest
}

func (f *HeartbeatRequest) MarshalNosTale() ([]byte, error) {
	return []byte{}, nil
}

func (f *HeartbeatRequest) UnmarshalNosTale(b []byte) error {
	return nil
}
