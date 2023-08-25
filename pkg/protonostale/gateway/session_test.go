package gateway

import (
	"bytes"
	"testing"

	eventing "go.shikanime.studio/elkia/pkg/api/eventing/v1alpha1"
	"go.shikanime.studio/elkia/pkg/protonostale"
	"google.golang.org/protobuf/proto"
)

func TestSessionDecodeSync(t *testing.T) {
	input := []byte("\x9f\xac\x64\xa0\x63\xeb\x67\x78\x63\x0e")
	expected := eventing.SyncRequest{
		Sequence: 59115,
		Code:     1098142510,
	}
	dec := NewSessionDecoder(bytes.NewReader(input))
	var result protonostale.SyncRequest
	if err := dec.Decode(&result); err != nil {
		t.Error(err)
	}
	if !proto.Equal(&expected, &result) {
		t.Errorf("Expected %v, got %v", expected.String(), result.String())
	}
}
