package encoding

import (
	"bytes"
	"io"
	"testing"
)

func TestSessionDecodeSyncFrame(t *testing.T) {
	input := []byte("\x96\xa5\xaa\xe0\x4f\x0e")
	expected := []byte("4352579 0 ;;")
	result, err := io.ReadAll(NewSessionReader(bytes.NewReader(input)))
	if err != nil {
		t.Errorf("Error reading line: %s", err)
	}
	if !bytes.Equal(expected, result) {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}
