package protonostale

import (
	"bytes"
	"testing"

	eventing "go.shikanime.studio/elkia/pkg/api/eventing/v1alpha1"
	"google.golang.org/protobuf/proto"
)

func TestLoginFrameUnmarshalNosTale(t *testing.T) {
	input := []byte("NoS0575 2503350 admin 9827F3538326B33722633327E4 006666A8\v0.9.3.3086")
	expected := AuthInteractRequest{
		&eventing.AuthInteractRequest{
			Request: &eventing.AuthInteractRequest_CreateLoginFlow{
				CreateLoginFlow: &eventing.CreateLoginFlowRequest{
					Identifier:    "admin",
					Password:      "s3cr3t",
					ClientVersion: "0.9.3+3086",
				},
			},
		},
	}
	var result AuthInteractRequest
	if err := result.UnmarshalNosTale(input); err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if !proto.Equal(&expected, &result) {
		t.Errorf("Expected %v, got %v", expected.String(), result.String())
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
	input := &CreateLoginFlowResponse{
		CreateLoginFlowResponse: &eventing.CreateLoginFlowResponse{
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

func TestSyncFrameUnmarshalNosTale(t *testing.T) {
	input := []byte("4349270 0 ;;")
	expected := eventing.SyncRequest{
		Sequence: 49270,
		Code:     0,
	}
	var result SyncRequest
	if err := result.UnmarshalNosTale(input); err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if !proto.Equal(&expected, &result) {
		t.Errorf("Expected %v, got %v", expected.String(), result.String())
	}
}

func TestIdentifierFrameUnmarshalNosTale(t *testing.T) {
	input := []byte("60471 ricofo8350@otanhome.com")
	expected := eventing.Identifier{
		Sequence: 60471,
		Value:    "ricofo8350@otanhome.com",
	}
	var result Identifier
	if err := result.UnmarshalNosTale(input); err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if !proto.Equal(&expected, &result) {
		t.Errorf("Expected %v, got %v", expected.String(), result.String())
	}
}

func TestPasswordFrameUnmarshalNosTale(t *testing.T) {
	input := []byte("60472 9hibwiwiG2e6Nr")
	expected := eventing.Password{
		Sequence: 60472,
		Value:    "9hibwiwiG2e6Nr",
	}
	var result Password
	if err := result.UnmarshalNosTale(input); err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if !proto.Equal(&expected, &result) {
		t.Errorf("Expected %v, got %v", expected.String(), result.String())
	}
}
