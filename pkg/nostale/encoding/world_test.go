package encoding

import (
	"bytes"
	"testing"
)

func TestWorldEncodingDecodeHandoffIdentifierFrame(t *testing.T) {
	input := []byte{
		198, 228, 203, 145, 70, 205, 214, 220, 208, 217, 208, 196, 7, 212, 73,
		255, 208, 203, 222, 209, 215, 208, 210, 218, 193, 112, 67, 220, 208,
		210, 63, 199, 228, 203, 161, 16, 72, 215, 214, 221, 200, 214, 200, 214,
		248, 193, 160, 65, 218, 193, 224, 66, 241, 205, 63,
	}
	expected := []byte("60471 ricofo8350@otanhome.com \xff")
	dec := NewDecoder(bytes.NewReader(input), WorldEncoding)
	result := make([]byte, WorldEncoding.DecodedLen(len(input)))
	if err := dec.Decode(&result); err != nil {
		t.Errorf("Error reading line: %s", err)
	}
	if !bytes.Equal(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestWorldEncodingDecodeHandoffPasswordFrame(t *testing.T) {
	input := []byte{
		199, 228, 203, 161, 16, 72, 215, 214, 221, 200, 214, 200, 214,
		248, 193, 160, 65, 218, 193, 224, 66, 241, 205, 63,
	}
	expected := []byte("60472 9hibwiwiG2e6Nr \xb1\x8d\xff")
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
		199, 205, 171, 241, 128, 63,
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
