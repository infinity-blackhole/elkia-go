package gateway

import (
	"bytes"
	"testing"

	"github.com/infinity-blackhole/elkia/pkg/nostale/auth"
)

// key:3324008792

func TestReaderReadSyncFrame(t *testing.T) {
	input := []byte{
		150, 156, 122, 80, 79, 14, 198, 205, 171, 145, 70, 205, 214, 220, 208,
		217, 208, 196, 7, 212, 73,
	}
	expected := []byte("4349270 0 ;;737:584-.37:83898 868 71;481.6")
	enc := auth.NewDecoder(NewSessionDecoding(), bytes.NewReader(input))
	result := make([]byte, len(expected))
	if err := enc.Decode(result); err != nil {
		t.Errorf("Error reading line: %s", err)
	}
	if !bytes.Equal(expected, result) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestReaderReadAuthHandoffPasswordFrame(t *testing.T) {
	input := []byte{
		208, 203, 222, 209, 215, 208, 210, 218, 193, 112, 67, 220, 208, 210,
		63, 199, 205, 171, 161, 16, 72, 215, 214, 221, 200, 214, 200, 214, 248,
		193, 160, 65, 218, 193, 224, 66, 241, 205, 63,
	}
	expected := []byte("49272 9hibwiwiG2e6Nr")
	dec := auth.NewDecoder(NewWorldFrameListEncoding(0), bytes.NewReader(input))
	result := make([]byte, len(expected))
	if err := dec.Decode(&result); err != nil {
		t.Errorf("Error reading line: %s", err)
	}
	if !bytes.Equal(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestReaderReadCodeDerivation(t *testing.T) {
	r := NewWorldFrameListEncoding(100)
	if r.mode != 1 {
		t.Errorf("Expected mode 74, got %d", r.mode)
	}
	if r.offset != 164 {
		t.Errorf("Expected offset 0, got %d", r.offset)
	}
	r = NewWorldFrameListEncoding(1)
	if r.mode != 0 {
		t.Errorf("Expected mode 0, got %d", r.mode)
	}
	if r.offset != 65 {
		t.Errorf("Expected offset 65, got %d", r.offset)
	}
}

func TestReaderReadeHeartbeatFrame(t *testing.T) {
	input := []byte{
		199, 205, 171, 241, 128, 63,
	}
	expected := []byte("49277 0")
	enc := auth.NewDecoder(NewWorldFrameListEncoding(0), bytes.NewReader(input))
	result := make([]byte, len(expected))
	if err := enc.Decode(&result); err != nil {
		t.Errorf("Error reading line: %s", err)
	}
	if !bytes.Equal(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestWriterWrite(t *testing.T) {
	input := "foo"
	expected := []byte{
		3, 153, 144, 144,
	}
	var result bytes.Buffer
	enc := auth.NewEncoder(NewSessionDecoding(), &result)
	if err := enc.Encode(&input); err != nil {
		t.Errorf("Error writing line: %s", err)
	}
	if !bytes.Equal(expected, result.Bytes()) {
		t.Errorf("Expected %v, got %v", expected, result.Bytes())
	}
}
