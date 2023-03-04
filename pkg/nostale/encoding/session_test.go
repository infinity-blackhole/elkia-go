package encoding

import (
	"bytes"
	"testing"
)

func TestSessionDecodeSyncFrame(t *testing.T) {
	input := []byte{
		150, 165, 170, 224, 79, 14,
	}
	expected := []byte("4352579 0 ;;")
	enc := NewDecoder(bytes.NewReader(input), SessionEncoding)
	result := make([]byte, SessionEncoding.DecodedLen(len(expected)))
	if err := enc.Decode(&result); err != nil {
		t.Errorf("Error reading line: %s", err)
	}
	if !bytes.Equal(expected, result) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}
