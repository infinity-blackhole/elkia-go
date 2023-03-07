package encoding

import (
	"bytes"
	"io"
	"testing"
)

func TestWorldReadIdentifierFrame(t *testing.T) {
	input := []byte(
		"\xc6\xe4\xcb\x91\x46\xcd\xd6\xdc\xd0\xd9\xd0\xc4\x07\xd4\x49\xff\xd0" +
			"\xcb\xde\xd1\xd7\xd0\xd2\xda\xc1\x70\x43\xdc\xd0\xd2\x3f\xc7\xe4" +
			"\xcb\xa1\x10\x48\xd7\xd6\xdd\xc8\xd6\xc8\xd6\xf8\xc1\xa0\x41\xda" +
			"\xc1\xe0\x42\xf1\xcd\x3f",
	)
	expected := []byte("60471 ricofo8350@otanhome.com\n60472 9hibwiwiG2e6Nr\n")
	result, err := io.ReadAll(
		NewWorldPackReader(NewWorldReader(bytes.NewReader(input), 0)),
	)
	if err != nil {
		t.Errorf("Error reading line: %s", err)
	}
	if !bytes.Equal(result, expected) {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

func TestWorldPasswordFrame(t *testing.T) {
	input := []byte(
		"\xc7\xe4\xcb\xa1\x10\x48\xd7\xd6\xdd\xc8\xd6\xc8\xd6\xf8\xc1\xa0\x41" +
			"\xda\xc1\xe0\x42\xf1\xcd\x3f",
	)
	expected := []byte("60472 9hibwiwiG2e6Nr\n")
	result, err := io.ReadAll(
		NewWorldPackReader(NewWorldReader(bytes.NewReader(input), 0)),
	)
	if err != nil {
		t.Errorf("Error reading line: %s", err)
	}
	if !bytes.Equal(result, expected) {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}
func TestWorldReadModeAndOffset(t *testing.T) {
	r := NewWorldReader(nil, 100)
	if r.mode != 1 {
		t.Errorf("Expected mode 74, got %d", r.mode)
	}
	if r.offset != 164 {
		t.Errorf("Expected offset 0, got %d", r.offset)
	}
	r = NewWorldReader(nil, 1)
	if r.mode != 0 {
		t.Errorf("Expected mode 0, got %d", r.mode)
	}
	if r.offset != 65 {
		t.Errorf("Expected offset 65, got %d", r.offset)
	}
}

func TestWorldReadHeartbeatFrame(t *testing.T) {
	input := []byte("\xc7\xcd\xab\xf1\x80\x3f")
	expected := []byte("49277 0\n")
	result, err := io.ReadAll(
		NewWorldPackReader(NewWorldReader(bytes.NewReader(input), 0)),
	)
	if err != nil {
		t.Errorf("Error reading line: %s", err)
	}
	if !bytes.Equal(result, expected) {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

func TestWorldWrite(t *testing.T) {
	input := []byte("foo\n")
	expected := []byte("\x03\x99\x90\x90\xff")
	var result bytes.Buffer
	w := NewWorldWriter(&result)
	if _, err := w.Write(input); err != nil {
		t.Errorf("Error writing line: %s", err)
	}
	if !bytes.Equal(expected, result.Bytes()) {
		t.Errorf("Expected %v, got %v", expected, result.Bytes())
	}
}
