package protonostale

import (
	"bufio"
	"bytes"
	"testing"

	eventing "github.com/infinity-blackhole/elkia/pkg/api/eventing/v1alpha1"
	"github.com/infinity-blackhole/elkia/pkg/nostale/simplesubtitution"
)

func TestParseAuthLoginEvent(t *testing.T) {
	input := []byte("2503350 admin 9827F3538326B33722633327E4 006666A8\v0.9.3.3086\n")
	expected := &eventing.AuthLoginEvent{
		Identifier:    "admin",
		Password:      "s3cr3t",
		ClientVersion: "0.9.3+3086",
	}
	result, err := NewAuthEventReader(bytes.NewBuffer(input)).ReadAuthLoginEvent()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if result.Identifier != expected.Identifier {
		t.Errorf("Expected %v, got %v", expected.Identifier, result.Identifier)
	}
	if result.Password != expected.Password {
		t.Errorf("Expected %v, got %v", expected.Password, result.Password)
	}
	if result.ClientVersion != expected.ClientVersion {
		t.Errorf("Expected %v, got %v", expected.ClientVersion, result.ClientVersion)
	}
}
func TestReadPassword(t *testing.T) {
	input := []byte("2EB6A196E4B60D96A9267E")
	expected := "admin"
	result, err := NewAuthEventReader(bytes.NewBuffer(input)).ReadPassword()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if result != expected {
		t.Errorf("Expected %v, got %v", expected, result)
	}

	input = []byte("1BE97B527A306B597A2")
	expected = "user"
	result, err = NewAuthEventReader(bytes.NewBuffer(input)).ReadPassword()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if result != expected {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestReadVersion(t *testing.T) {
	input := []byte{48, 48, 57, 55, 54, 55, 69, 67, 11, 48, 46, 57, 46, 51, 46, 51, 48, 56, 54, 10}
	expected := "0.9.3+3086"
	result, err := NewAuthEventReader(bytes.NewBuffer(input)).ReadVersion()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if result != expected {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestWriteGatewayListEvent(t *testing.T) {
	input := &eventing.GatewayListEvent{
		Code: 1,
		Gateways: []*eventing.Gateway{
			{
				Host:      "127.0.0.1",
				Port:      "4124",
				Weight:    0,
				WorldId:   1,
				ChannelId: 1,
				WorldName: "Test",
			},
			{
				Host:      "127.0.0.1",
				Port:      "4125",
				Weight:    0,
				WorldId:   1,
				ChannelId: 2,
				WorldName: "Test",
			},
		},
	}
	var expected bytes.Buffer
	if _, err := simplesubtitution.NewWriter(
		bufio.NewWriter(&expected),
	).Write([]byte(
		"NsTeST 1 127.0.0.1:4124:0:1.1.Test 127.0.0.1:4125:0:1.2.Test -1:-1:-1:10000.10000.1",
	)); err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	var result bytes.Buffer
	if err := NewAuthWriter(bufio.NewWriter(&result)).
		WriteGatewayListEvent(input); err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if !bytes.Equal(result.Bytes(), expected.Bytes()) {
		t.Errorf("Expected %v, got %v", expected.Bytes(), result.Bytes())
	}
}
