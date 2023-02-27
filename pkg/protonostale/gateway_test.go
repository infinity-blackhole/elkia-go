package protonostale

import (
	"testing"

	eventing "github.com/infinity-blackhole/elkia/pkg/api/eventing/v1alpha1"
)

func TestParseAuthHandoffSyncEvent(t *testing.T) {
	input := "4349270 0 ;;737:584-.37:83898 868 71;481.6; "
	expected := &eventing.AuthHandoffSyncEvent{
		Key: 0,
	}
	result, err := ParseAuthHandoffSyncEvent(input)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if result.Key != expected.Key {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}
