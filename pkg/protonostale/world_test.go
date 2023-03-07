package protonostale

import (
	"testing"

	eventing "github.com/infinity-blackhole/elkia/pkg/api/eventing/v1alpha1"
)

func TestSyncFrameUnmarshalNosTale(t *testing.T) {
	input := []byte("4349270 0 ;;")
	expected := &eventing.SyncFrame{
		Code: 0,
	}
	var result SyncFrame
	if err := result.UnmarshalNosTale(input); err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if result.Code != expected.Code {
		t.Errorf("Expected %v, got %v", expected, result.String())
	}
}

func TestIdentifierFrameUnmarshalNosTale(t *testing.T) {
	input := []byte("60471 ricofo8350@otanhome.com")
	expected := &eventing.IdentifierFrame{
		Sequence:   60471,
		Identifier: "ricofo8350@otanhome.com",
	}
	var result IdentifierFrame
	if err := result.UnmarshalNosTale(input); err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if result.Sequence != expected.Sequence {
		t.Errorf("Expected %v, got %v", expected, result.String())
	}
	if result.Identifier != expected.Identifier {
		t.Errorf("Expected %v, got %v", expected, result.String())
	}
}

func TestPasswordFrameUnmarshalNosTale(t *testing.T) {
	input := []byte("60472 9hibwiwiG2e6Nr")
	expected := &eventing.PasswordFrame{
		Sequence: 60472,
		Password: "9hibwiwiG2e6Nr",
	}
	var result PasswordFrame
	if err := result.UnmarshalNosTale(input); err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if result.Sequence != expected.Sequence {
		t.Errorf("Expected %v, got %v", expected, result.String())
	}
	if result.Password != expected.Password {
		t.Errorf("Expected %v, got %v", expected, result.String())
	}
}
