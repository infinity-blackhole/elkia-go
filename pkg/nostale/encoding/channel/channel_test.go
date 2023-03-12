package encoding

import (
	"bytes"
	"testing"
)

func TestWorldReadIdentifierFrame(t *testing.T) {
	input := []byte(
		"\xc6\xe4\xcb\x91\x46\xcd\xd6\xdc\xd0\xd9\xd0\xc4\x07\xd4\x49\xff\xd0" +
			"\xcb\xde\xd1\xd7\xd0\xd2\xda\xc1\x70\x43\xdc\xd0\xd2\x3f\xc7\xe4" +
			"\xcb\xa1\x10\x48\xd7\xd6\xdd\xc8\xd6\xc8\xd6\xf8\xc1\xa0\x41\xda" +
			"\xc1\xe0\x42\xf1\xcd",
	)
	expected := [][]byte{[]byte("60471 ricofo8350@otanhome.com"), []byte("60472 9hibwiwiG2e6Nr")}
	for i, input := range bytes.Split(input, []byte("\x3f")) {
		result := make([]byte, DecodeFrameMaxLen(len(input)))
		n := DecodeFrame(result, input, 0)
		if !bytes.Equal(expected[i], result[:n]) {
			t.Errorf("Expected %s, got %s", expected, result[:n])
		}
	}
}

func TestWorldPasswordFrame(t *testing.T) {
	input := []byte(
		"\xc7\xe4\xcb\xa1\x10\x48\xd7\xd6\xdd\xc8\xd6\xc8\xd6\xf8\xc1\xa0\x41" +
			"\xda\xc1\xe0\x42\xf1\xcd",
	)
	expected := []byte("60472 9hibwiwiG2e6Nr")
	result := make([]byte, DecodeFrameMaxLen(len(input)))
	n := DecodeFrame(result, input, 0)
	if !bytes.Equal(expected, result[:n]) {
		t.Errorf("Expected %s, got %s", expected, result[:n])
	}
}
func TestWorldReadModeAndOffset(t *testing.T) {
	mode, offset := DecodeKey(100)
	if mode != 1 {
		t.Errorf("Expected mode 74, got %d", mode)
	}
	if offset != 164 {
		t.Errorf("Expected offset 0, got %d", offset)
	}
	mode, offset = DecodeKey(1)
	if mode != 0 {
		t.Errorf("Expected mode 0, got %d", mode)
	}
	if offset != 65 {
		t.Errorf("Expected offset 65, got %d", offset)
	}
}

func TestWorldReadHeartbeatFrame(t *testing.T) {
	input := []byte("\xc7\xcd\xab\xf1\x80")
	expected := []byte("49277 0")
	result := make([]byte, DecodeFrameMaxLen(len(input)))
	n := DecodeFrame(result, input, 0)
	if !bytes.Equal(expected, result[:n]) {
		t.Errorf("Expected %s, got %s", expected, result[:n])
	}
}

func TestWorldWrite(t *testing.T) {
	input := []byte("foo")
	expected := []byte("\x03\x99\x90\x90")
	result := make([]byte, EncodeFrameMaxLen(len(input)))
	n := EncodeFrame(result, input)
	if !bytes.Equal(expected, result[:n]) {
		t.Errorf("Expected %v, got %v", expected, result[:n])
	}
}
