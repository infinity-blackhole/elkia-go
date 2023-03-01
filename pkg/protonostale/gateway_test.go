package protonostale

import (
	"testing"

	eventing "github.com/infinity-blackhole/elkia/pkg/api/eventing/v1alpha1"
)

func TestDecodeAuthHandoffSyncFrame(t *testing.T) {
	input := []byte("4349270 0 ;;737:584-.37:83898 868 71;481.6; ")
	expected := &eventing.AuthHandoffSyncFrame{
		Code: 0,
	}
	result, err := DecodeAuthHandoffSyncFrame(input)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if result.Code != expected.Code {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}
