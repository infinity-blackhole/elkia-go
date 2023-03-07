package protonostale

import (
	"bytes"
	"testing"

	eventing "github.com/infinity-blackhole/elkia/pkg/api/eventing/v1alpha1"
)

func TestAuthLoginFrameUnmarshalNosTale(t *testing.T) {
	input := []byte("2503350 admin 9827F3538326B33722633327E4 006666A8\v0.9.3.3086")
	expected := &eventing.AuthLoginFrame{
		Identifier:    "admin",
		Password:      "s3cr3t",
		ClientVersion: "0.9.3+3086",
	}
	var result AuthLoginFrame
	if err := result.UnmarshalNosTale(input); err != nil {
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

func TestDecodePassword(t *testing.T) {
	input := []byte("2EB6A196E4B60D96A9267E")
	expected := "admin"
	result, err := DecodePassword(input)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if result != expected {
		t.Errorf("Expected %v, got %v", expected, result)
	}

	input = []byte("1BE97B527A306B597A2")
	expected = "user"
	result, err = DecodePassword(input)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if result != expected {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestDecodeClientVersion(t *testing.T) {
	input := []byte("006666A8\v0.9.3.3086")
	expected := "0.9.3+3086"
	result, err := DecodeClientVersion(input)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if result != expected {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestEndpointListFrameUnmarshalNosTale(t *testing.T) {
	input := &EndpointListFrame{
		EndpointListFrame: eventing.EndpointListFrame{
			Code: 1,
			Endpoints: []*eventing.Endpoint{
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
		},
	}
	expected := []byte("NsTeST 1 127.0.0.1:4124:0:1.1.Test 127.0.0.1:4125:0:1.2.Test -1:-1:-1:10000.10000.1\n")
	result, err := input.MarshalNosTale()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if !bytes.Equal(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}
