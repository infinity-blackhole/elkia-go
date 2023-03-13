package gateway

import (
	"bytes"
	"testing"

	eventing "github.com/infinity-blackhole/elkia/pkg/api/eventing/v1alpha1"
	"github.com/infinity-blackhole/elkia/pkg/protonostale"
	"google.golang.org/protobuf/proto"
)

func TestChannelReadIdentifierFrame(t *testing.T) {
	input := []byte(
		"\xc6\xe4\xcb\x91\x46\xcd\xd6\xdc\xd0\xd9\xd0\xc4\x07\xd4\x49\xff\xd0" +
			"\xcb\xde\xd1\xd7\xd0\xd2\xda\xc1\x70\x43\xdc\xd0\xd2\x3f\xc7\xe4" +
			"\xcb\xa1\x10\x48\xd7\xd6\xdd\xc8\xd6\xc8\xd6\xf8\xc1\xa0\x41\xda" +
			"\xc1\xe0\x42\xf1\xcd",
	)
	expected := []protonostale.ChannelInteractRequest{
		{
			ChannelInteractRequest: &eventing.ChannelInteractRequest{
				Sequence: 60471,
				Payload: &eventing.ChannelInteractRequest_RawFrame{
					RawFrame: []byte("ricofo8350@otanhome.com"),
				},
			},
		},
		{
			ChannelInteractRequest: &eventing.ChannelInteractRequest{
				Sequence: 60472,
				Payload: &eventing.ChannelInteractRequest_RawFrame{
					RawFrame: []byte("9hibwiwiG2e6Nr"),
				},
			},
		},
	}
	for i, input := range bytes.Split(input, []byte("\x3f")) {
		enc := NewChannelDecoder(bytes.NewReader(input), 0)
		var result protonostale.ChannelInteractRequest
		if err := enc.Decode(&result); err != nil {
			t.Fatal(err)
		}
		if !proto.Equal(expected[i], result) {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	}
}

func TestChannelPasswordFrame(t *testing.T) {
	input := []byte(
		"\xc7\xe4\xcb\xa1\x10\x48\xd7\xd6\xdd\xc8\xd6\xc8\xd6\xf8\xc1\xa0\x41" +
			"\xda\xc1\xe0\x42\xf1\xcd",
	)
	expected := protonostale.ChannelInteractRequest{
		ChannelInteractRequest: &eventing.ChannelInteractRequest{
			Sequence: 60472,
			Payload: &eventing.ChannelInteractRequest_RawFrame{
				RawFrame: []byte("9hibwiwiG2e6Nr"),
			},
		},
	}
	var result protonostale.ChannelInteractRequest
	enc := NewChannelDecoder(bytes.NewReader(input), 0)
	if err := enc.Decode(&result); err != nil {
		t.Fatal(err)
	}
	if !proto.Equal(expected, result) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}
func TestChannelReadModeAndOffset(t *testing.T) {
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

func TestChannelReadHeartbeatFrame(t *testing.T) {
	input := []byte("\xc7\xcd\xab\xf1\x80")
	expected := protonostale.ChannelInteractRequest{
		ChannelInteractRequest: &eventing.ChannelInteractRequest{
			Sequence: 49277,
			Payload:  &eventing.ChannelInteractRequest_HeartbeatFrame{},
		},
	}
	var result protonostale.ChannelInteractRequest
	enc := NewChannelDecoder(bytes.NewReader(input), 0)
	if err := enc.Decode(&result); err != nil {
		t.Fatal(err)
	}
	if !proto.Equal(expected, result) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestChannelWrite(t *testing.T) {
	input := []byte("foo")
	expected := []byte("\x03\x99\x90\x90")
	var buff bytes.Buffer
	enc := NewEncoder(&buff)
	if err := enc.Encode(input); err != nil {
		t.Fatal(err)
	}
	result := buff.Bytes()
	if !bytes.Equal(expected, result) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}
