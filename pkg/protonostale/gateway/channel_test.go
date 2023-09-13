package gateway

import (
	"bytes"
	"testing"

	eventingpb "go.shikanime.studio/elkia/pkg/api/eventing/v1alpha1"
	fleetpb "go.shikanime.studio/elkia/pkg/api/fleet/v1alpha1"
	"go.shikanime.studio/elkia/pkg/protonostale"
	"google.golang.org/protobuf/proto"
)

func TestChannelDecodeIdentifierCommand(t *testing.T) {
	input := []byte(
		"\xc6\xe4\xcb\x91\x46\xcd\xd6\xdc\xd0\xd9\xd0\xc4\x07\xd4\x49\xff\xd0" +
			"\xcb\xde\xd1\xd7\xd0\xd2\xda\xc1\x70\x43\xdc\xd0\xd2\x3f\xc7\xe4" +
			"\xcb\xa1\x10\x48\xd7\xd6\xdd\xc8\xd6\xc8\xd6\xf8\xc1\xa0\x41\xda" +
			"\xc1\xe0\x42\xf1\xcd",
	)
	expectedIdentifier := fleetpb.IdentifierCommand{
		Sequence:   60471,
		Identifier: "ricofo8350@otanhome.com",
	}
	expectedPassword := fleetpb.PasswordCommand{
		Sequence: 60472,
		Password: "9hibwiwiG2e6Nr",
	}
	dec := NewChannelDecoder(bytes.NewReader(input), 0)
	var resultIdentifier protonostale.IdentifierCommand
	if err := dec.Decode(&resultIdentifier); err != nil {
		t.Fatal(err)
	}
	if !proto.Equal(&expectedIdentifier, &resultIdentifier) {
		t.Errorf("Expected %v, got %v", expectedIdentifier.String(), resultIdentifier.String())
	}
	var resultPassword protonostale.PasswordCommand
	if err := dec.Decode(&resultPassword); err != nil {
		t.Fatal(err)
	}
	if !proto.Equal(&expectedPassword, &resultPassword) {
		t.Errorf("Expected %v, got %v", expectedPassword.String(), resultPassword.String())
	}
}

func TestChannelDecodePasswordCommand(t *testing.T) {
	input := []byte(
		"\xc7\xe4\xcb\xa1\x10\x48\xd7\xd6\xdd\xc8\xd6\xc8\xd6\xf8\xc1\xa0\x41" +
			"\xda\xc1\xe0\x42\xf1\xcd",
	)
	expected := protonostale.PasswordCommand{
		PasswordCommand: &fleetpb.PasswordCommand{
			Sequence: 60472,
			Password: "9hibwiwiG2e6Nr",
		},
	}
	var result protonostale.PasswordCommand
	enc := NewChannelDecoder(bytes.NewReader(input), 0)
	if err := enc.Decode(&result); err != nil {
		t.Fatal(err)
	}
	if !proto.Equal(&expected, &result) {
		t.Errorf("Expected %v, got %v", expected.String(), result.String())
	}
}

func TestChannelDecodeModeAndOffset(t *testing.T) {
	r := NewPackedChannelScanner(nil, 100)
	if r.mode != 1 {
		t.Errorf("Expected mode 74, got %d", r.mode)
	}
	if r.offset != 164 {
		t.Errorf("Expected offset 0, got %d", r.offset)
	}
	r = NewPackedChannelScanner(nil, 1)
	if r.mode != 0 {
		t.Errorf("Expected mode 0, got %d", r.mode)
	}
	if r.offset != 65 {
		t.Errorf("Expected offset 65, got %d", r.offset)
	}
}

func TestChannelDecodeHeartbeatCommand(t *testing.T) {
	input := []byte("\xc7\xcd\xab\xf1\x80")
	expected := protonostale.ClientCommand{
		ClientCommand: &eventingpb.ClientCommand{
			Sequence: 49277,
			Command:  &eventingpb.ClientCommand_Heartbeat{},
		},
	}
	var result protonostale.ClientCommand
	enc := NewChannelDecoder(bytes.NewReader(input), 0)
	if err := enc.Decode(&result); err != nil {
		t.Fatal(err)
	}
	if !proto.Equal(&expected, &result) {
		t.Errorf("Expected %v, got %v", expected.String(), result.String())
	}
}

func TestChannelWrite(t *testing.T) {
	input := []byte("foo")
	expected := []byte("\x03\x99\x90\x90\x19")
	var buff bytes.Buffer
	enc := NewWriter(&buff)
	if err := enc.WriteCommand(input); err != nil {
		t.Fatal(err)
	}
	result := buff.Bytes()
	if !bytes.Equal(expected, result) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}
