package auth

import (
	"bytes"
	"testing"

	eventing "go.shikanime.studio/elkia/pkg/api/eventing/v1alpha1"
	"go.shikanime.studio/elkia/pkg/protonostale"
)

func TestAuthEncodingDecodeCommand(t *testing.T) {
	input := []byte(
		"\x9c\xbb\x9f\x02\x05\x03\x05\xf2\xff\xff\x04\x05\xff\x06\x0a\xf2\xc0" +
			"\xb9\xaf\xbb\xb4\xbb\x0a\xff\x05\x02\x92\xbb\xc6\xb1\xbc\xba\xbb" +
			"\xbd\xb5\xfc\xaf\xbb\xbd\xf2\x01\x05\x01\xff\x01\x09\x05\x04\x90" +
			"\x0a\xff\x04\x03\x09\x05\x04\xff\x00\x06\x03\xff\x03\x01\x04\x05" +
			"\x09\xff\x03\x06\x03\x06\x04\x05\x09\x00\x06\x05\x03\x06\xff\x90" +
			"\x00\x01\x04\x96\x05\x00\xff\x94\x04\x05\x06\x0a\x95\x00\x03\x90" +
			"\x00\xf2\x02\x02\x04\x94\x03\x06\x06\x04\xd7\x02\xfc\x09\xfc\xff" +
			"\xfc\xff\x02\x0a\x04\xd8",
	)
	expected := eventing.HandoffFlowRequest{
		Identifier:    "ricofo8350@otanhome.com",
		Password:      "9hibwiwiG2e6Nr",
		ClientVersion: "0.9.3+3086",
	}
	var result protonostale.AuthInteractRequest
	if err := NewDecoder(bytes.NewReader(input)).Decode(&result); err != nil {
		t.Fatal(err)
	}
	if payload, ok := result.Payload.(*eventing.AuthInteractRequest_HandoffFlowRequest); ok {
		if expected.Identifier != payload.HandoffFlowRequest.Identifier {
			t.Errorf("Expected %v, got %v", expected.Identifier, payload.HandoffFlowRequest.Identifier)
		}
		if expected.Password != payload.HandoffFlowRequest.Password {
			t.Errorf("Expected %v, got %v", expected.Password, payload.HandoffFlowRequest.Password)
		}
		if expected.ClientVersion != payload.HandoffFlowRequest.ClientVersion {
			t.Errorf("Expected %v, got %v", expected.ClientVersion, payload.HandoffFlowRequest.ClientVersion)
		}
	}
}

func TestAuthEncodingEncodeCommand(t *testing.T) {
	input := protonostale.InfoEvent{
		InfoEvent: &eventing.InfoEvent{
			Content: "fail Hello. This is a basic test",
		},
	}
	expected := []byte(
		"\x78\x7d\x75\x7e\x2f\x75\x70\x78\x7b\x2f\x57\x74\x7b\x7b\x7e\x3d\x2f" +
			"\x63\x77\x78\x82\x2f\x78\x82\x2f\x70\x2f\x71\x70\x82\x78\x72\x2f" +
			"\x83\x74\x82\x83\x19",
	)
	var buff bytes.Buffer
	if err := NewEncoder(&buff).Encode(&input); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(expected, buff.Bytes()) {
		t.Errorf("Expected %v, got %v", expected, buff.Bytes())
	}
}
