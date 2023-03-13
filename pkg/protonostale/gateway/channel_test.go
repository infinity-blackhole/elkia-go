package gateway

import (
	"bytes"
	"testing"

	eventing "github.com/infinity-blackhole/elkia/pkg/api/eventing/v1alpha1"
	"github.com/infinity-blackhole/elkia/pkg/protonostale"
	"google.golang.org/protobuf/proto"
)

func TestChannelDecodeIdentifierFrame(t *testing.T) {
	input := []byte(
		"\xc6\xe4\xcb\x91\x46\xcd\xd6\xdc\xd0\xd9\xd0\xc4\x07\xd4\x49\xff\xd0" +
			"\xcb\xde\xd1\xd7\xd0\xd2\xda\xc1\x70\x43\xdc\xd0\xd2\x3f\xc7\xe4" +
			"\xcb\xa1\x10\x48\xd7\xd6\xdd\xc8\xd6\xc8\xd6\xf8\xc1\xa0\x41\xda" +
			"\xc1\xe0\x42\xf1\xcd",
	)
	expectedIdentifier := eventing.IdentifierFrame{
		Sequence:   60471,
		Identifier: "ricofo8350@otanhome.com",
	}
	expectedPassword := eventing.PasswordFrame{
		Sequence: 60472,
		Password: "9hibwiwiG2e6Nr",
	}
	dec := NewChannelDecoder(bytes.NewReader(input), 0)
	var resultIdentifier protonostale.IdentifierFrame
	if err := dec.Decode(&resultIdentifier); err != nil {
		t.Fatal(err)
	}
	if !proto.Equal(&expectedIdentifier, &resultIdentifier) {
		t.Errorf("Expected %v, got %v", expectedIdentifier.String(), resultIdentifier.String())
	}
	var resultPassword protonostale.PasswordFrame
	if err := dec.Decode(&resultPassword); err != nil {
		t.Fatal(err)
	}
	if !proto.Equal(&expectedPassword, &resultPassword) {
		t.Errorf("Expected %v, got %v", expectedPassword.String(), resultPassword.String())
	}
}

func TestChannelDecodePasswordFrame(t *testing.T) {
	input := []byte(
		"\xc7\xe4\xcb\xa1\x10\x48\xd7\xd6\xdd\xc8\xd6\xc8\xd6\xf8\xc1\xa0\x41" +
			"\xda\xc1\xe0\x42\xf1\xcd",
	)
	expected := protonostale.PasswordFrame{
		PasswordFrame: eventing.PasswordFrame{
			Sequence: 60472,
			Password: "9hibwiwiG2e6Nr",
		},
	}
	var result protonostale.PasswordFrame
	enc := NewSessionDecoder(bytes.NewReader(input))
	if err := enc.Decode(&result); err != nil {
		t.Fatal(err)
	}
	if !proto.Equal(&expected, &result) {
		t.Errorf("Expected %v, got %v", expected.String(), result.String())
	}
}

func TestChannelDecodeModeAndOffset(t *testing.T) {
	r := NewChannelScanner(nil, 100)
	if r.mode != 1 {
		t.Errorf("Expected mode 74, got %d", r.mode)
	}
	if r.offset != 164 {
		t.Errorf("Expected offset 0, got %d", r.offset)
	}
	r = NewChannelScanner(nil, 1)
	if r.mode != 0 {
		t.Errorf("Expected mode 0, got %d", r.mode)
	}
	if r.offset != 65 {
		t.Errorf("Expected offset 65, got %d", r.offset)
	}
}

func TestChannelDecodeHeartbeatFrame(t *testing.T) {
	input := []byte("\xc7\xcd\xab\xf1\x80")
	expected := protonostale.ChannelFrame{
		ChannelFrame: eventing.ChannelFrame{
			Sequence: 49277,
			Payload:  &eventing.ChannelFrame_HeartbeatFrame{},
		},
	}
	var result protonostale.ChannelFrame
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
	if err := enc.WriteFrame(input); err != nil {
		t.Fatal(err)
	}
	result := buff.Bytes()
	if !bytes.Equal(expected, result) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}
