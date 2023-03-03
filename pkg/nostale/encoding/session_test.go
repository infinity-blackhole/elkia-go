package encoding

import (
	"bytes"
	"testing"
)

func TestSessionDecodeSyncFrame(t *testing.T) {
	input := []byte{
		150, 156, 122, 80, 79, 14, 198, 205, 171, 145, 70, 205, 214, 220, 208,
		217, 208, 196, 7, 212, 73, 255,
	}
	expected := []byte("4349270 0 ;;737:584-.37:83898 868 71;481.6; ")
	enc := NewDecoder(bytes.NewReader(input), SessionEncoding)
	result := make([]byte, SessionEncoding.DecodedLen(len(expected)))
	if err := enc.Decode(&result); err != nil {
		t.Errorf("Error reading line: %s", err)
	}
	if !bytes.Equal(expected, result) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}
