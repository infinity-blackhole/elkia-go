package gateway

import (
	"bytes"
	"testing"

	eventing "github.com/infinity-blackhole/elkia/pkg/api/eventing/v1alpha1"
	"github.com/infinity-blackhole/elkia/pkg/protonostale"
	"google.golang.org/protobuf/proto"
)

func TestSessionDecodeSyncFrame(t *testing.T) {
	input := []byte("\x96\xa5\xaa\xe0\x4f\x0e")
	expected := protonostale.ChannelInteractRequest{
		ChannelInteractRequest: &eventing.ChannelInteractRequest{
			Sequence: 4352579,
			Payload: &eventing.ChannelInteractRequest_RawFrame{
				RawFrame: []byte("0 ;;"),
			},
		},
	}
	var result protonostale.ChannelInteractRequest
	enc := NewSessionDecoder(bytes.NewReader(input))
	if err := enc.Decode(&result); err != nil {
		t.Fatal(err)
	}
	if !proto.Equal(expected, result) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}
