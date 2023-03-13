package protonostale

import (
	"testing"

	eventing "github.com/infinity-blackhole/elkia/pkg/api/eventing/v1alpha1"
	"google.golang.org/protobuf/proto"
)

func TestSyncFrameUnmarshalNosTale(t *testing.T) {
	input := []byte("4349270 0 ;;")
	expected := &ChannelInteractRequest{
		&eventing.ChannelInteractRequest{
			Sequence: 4349270,
			Payload: &eventing.ChannelInteractRequest_SyncFrame{
				SyncFrame: &eventing.SyncFrame{
					Code: 0,
				},
			},
		},
	}
	var result ChannelInteractRequest
	if err := result.UnmarshalNosTale(input); err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if !proto.Equal(expected, result) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestIdentifierFrameUnmarshalNosTale(t *testing.T) {
	input := []byte("60471 ricofo8350@otanhome.com")
	expected := &ChannelInteractRequest{
		&eventing.ChannelInteractRequest{
			Sequence: 60471,
			Payload: &eventing.ChannelInteractRequest_IdentifierFrame{
				IdentifierFrame: &eventing.IdentifierFrame{
					Identifier: "ricofo8350@otanhome.com",
				},
			},
		},
	}
	var result ChannelInteractRequest
	if err := result.UnmarshalNosTale(input); err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if !proto.Equal(expected, result) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestPasswordFrameUnmarshalNosTale(t *testing.T) {
	input := []byte("60472 9hibwiwiG2e6Nr")
	expected := &ChannelInteractRequest{
		&eventing.ChannelInteractRequest{
			Sequence: 60472,
			Payload: &eventing.ChannelInteractRequest_PasswordFrame{
				PasswordFrame: &eventing.PasswordFrame{
					Password: "9hibwiwiG2e6Nr",
				},
			},
		},
	}
	var result ChannelInteractRequest
	if err := result.UnmarshalNosTale(input); err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if !proto.Equal(expected, result) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}
