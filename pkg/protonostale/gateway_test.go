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
