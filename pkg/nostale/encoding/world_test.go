package encoding

import (
	"bytes"
	"testing"
)

func TestWorldEncodingDecodeHandoffPasswordFrame(t *testing.T) {
	input := []byte{
		208, 203, 222, 209, 215, 208, 210, 218, 193, 112, 67, 220, 208, 210,
		63, 199, 205, 171, 161, 16, 72, 215, 214, 221, 200, 214, 200, 214, 248,
		193, 160, 65, 218, 193, 224, 66, 241, 205, 63, 10,
	}
	// The payload is "49272 9hibwiwiG2e6Nr", the rest is the keep alive
	// and protocol garbage.
	expected := []byte("475n5 5355-564 .com49272 9hibwiwiG2e6Nr\xca")
	dec := NewDecoder(bytes.NewReader(input), WorldEncoding)
	result := make([]byte, WorldEncoding.DecodedLen(len(input)))
	if err := dec.Decode(&result); err != nil {
		t.Errorf("Error reading line: %s", err)
	}
	if !bytes.Equal(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestWorldEncodingDecodeModeAndOffset(t *testing.T) {
	r := WorldEncoding.WithKey(100)
	if r.mode() != 1 {
		t.Errorf("Expected mode 74, got %d", r.mode())
	}
	if r.offset() != 164 {
		t.Errorf("Expected offset 0, got %d", r.offset())
	}
	r = WorldEncoding.WithKey(1)
	if r.mode() != 0 {
		t.Errorf("Expected mode 0, got %d", r.mode())
	}
	if r.offset() != 65 {
		t.Errorf("Expected offset 65, got %d", r.offset())
	}
}

func TestWorldEncodingDecodeHeartbeatFrame(t *testing.T) {
	input := []byte{
		199, 205, 171, 241, 128, 63, 10,
	}
	expected := []byte("49277 0")
	dec := NewDecoder(bytes.NewReader(input), WorldEncoding)
	result := make([]byte, WorldEncoding.DecodedLen(len(input)))
	if err := dec.Decode(&result); err != nil {
		t.Errorf("Error reading line: %s", err)
	}
	if !bytes.Equal(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestWorldEncodingEncode(t *testing.T) {
	input := "foo"
	expected := []byte{3, 153, 144, 144, 10}
	var result bytes.Buffer
	enc := NewEncoder(&result, WorldEncoding)
	if err := enc.Encode(&input); err != nil {
		t.Errorf("Error writing line: %s", err)
	}
	if !bytes.Equal(expected, result.Bytes()) {
		t.Errorf("Expected %v, got %v", expected, result.Bytes())
	}
}
