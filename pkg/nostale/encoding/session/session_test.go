package encoding

import (
	"bytes"
	"testing"
)

func TestSessionDecodeSyncFrame(t *testing.T) {
	input := []byte("\x96\xa5\xaa\xe0\x4f\x0e")
	expected := []byte("4352579 0 ;;")
	result := make([]byte, DecodeFrameMaxLen(len(input)))
	n := DecodeFrame(result, input)
	if !bytes.Equal(expected, result[:n]) {
		t.Errorf("Expected %s, got %s", expected, result[:n])
	}
}
