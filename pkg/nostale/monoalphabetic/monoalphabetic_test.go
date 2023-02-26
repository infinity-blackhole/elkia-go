package monoalphabetic

import (
	"bufio"
	"bytes"
	"testing"
)

func TestReaderRead(t *testing.T) {
	input := []byte{
		150, 153, 171, 192, 79, 14, 198, 202, 220, 1, 70, 205, 214, 220, 208,
		217, 208, 196, 7, 212, 73, 255,
	}
	expected := []byte{
		135, 138, 156, 177, 64, 255, 183, 187, 205, 242, 55, 190, 199, 205,
		193, 202, 193, 181, 248, 197, 58, 240,
	}
	r := NewReader(bufio.NewReader(bytes.NewReader(input)))
	result := make([]byte, len(expected))
	n, err := r.Read(result)
	if err != nil {
		t.Errorf("Error reading line: %s", err)
	}
	if n != len(expected) {
		t.Errorf("Expected %d bytes, got %d", len(expected), n)
	}
	if !bytes.Equal(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}
