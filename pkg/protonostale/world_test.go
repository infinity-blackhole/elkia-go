package protonostale

import (
	"testing"

	eventing "github.com/infinity-blackhole/elkia/pkg/api/eventing/v1alpha1"
)

func TestSyncFrameUnmarshalNosTale(t *testing.T) {
	input := []byte("4349270 0 ;;737:584-.37:83898 868 71;481.6; ")
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

func TestHandoffFrameUnmarshallNosTale(t *testing.T) {
	input := []byte("60471 ricofo8350@otanhome.com 60472 9hibwiwiG2e6Nr \x02\xb1\x8d\xff\xca")
	expected := &eventing.HandoffFrame{
		IdentifierFrame: &eventing.HandoffIdentifierFrame{
			Sequence:   60471,
			Identifier: "ricofo8350@otanhome.com",
		},
		PasswordFrame: &eventing.HandoffPasswordFrame{
			Sequence: 60472,
			Password: "9hibwiwiG2e6Nr",
		},
	}
	var result HandoffFrame
	if err := result.UnmarshalNosTale(input); err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if result.IdentifierFrame.Sequence != expected.IdentifierFrame.Sequence {
		t.Errorf("Expected %v, got %v", expected, result.String())
	}
	if result.IdentifierFrame.Identifier != expected.IdentifierFrame.Identifier {
		t.Errorf("Expected %v, got %v", expected, result.String())
	}
	if result.PasswordFrame.Sequence != expected.PasswordFrame.Sequence {
		t.Errorf("Expected %v, got %v", expected, result.String())
	}
	if result.PasswordFrame.Password != expected.PasswordFrame.Password {
		t.Errorf("Expected %v, got %v", expected, result.String())
	}
}
