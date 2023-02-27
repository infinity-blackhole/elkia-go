package gateway

import (
	"bufio"
	"bytes"
	"testing"
	"testing/iotest"
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
	r := iotest.NewReadLogger(
		t.Name(),
		NewReader(bufio.NewReader(bytes.NewReader(input))),
	)
	result := make([]byte, len(expected))
	n, err := r.Read(result)
	if err != nil {
		t.Errorf("Error reading line: %s", err)
	}
	if !bytes.Equal(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
	if n != len(expected) {
		t.Errorf("Expected %d bytes, got %d", len(expected), n)
	}
}

// key:3324008792

func TestReaderReadSyncEvent(t *testing.T) {
	input := []byte{
		150, 156, 122, 80, 79, 14, 198, 205, 171, 145, 70, 205, 214, 220, 208,
		217, 208, 196, 7, 212, 73, 255,
	}
	expected := []byte("0")
	r := iotest.NewReadLogger(
		t.Name(),
		NewReader(bufio.NewReader(bytes.NewReader(input))),
	)
	result := make([]byte, len(expected))
	n, err := r.Read(result)
	if err != nil {
		t.Errorf("Error reading line: %s", err)
	}
	if !bytes.Equal(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
	if n != len(expected) {
		t.Errorf("Expected %d bytes, got %d", len(expected), n)
	}
}

func TestReaderReadAuthHandoffPasswordEvent(t *testing.T) {
	input := []byte{
		208, 203, 222, 209, 215, 208, 210, 218, 193, 112, 67, 220, 208, 210,
		63, 199, 205, 171, 161, 16, 72, 215, 214, 221, 200, 214, 200, 214, 248,
		193, 160, 65, 218, 193, 224, 66, 241, 205, 63,
	}
	expected := []byte("49272 9hibwiwiG2e6Nr")
	r := iotest.NewReadLogger(
		t.Name(),
		NewReader(bufio.NewReader(bytes.NewReader(input))),
	)
	result := make([]byte, len(expected))
	n, err := r.Read(result)
	if err != nil {
		t.Errorf("Error reading line: %s", err)
	}
	if !bytes.Equal(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
	if n != len(expected) {
		t.Errorf("Expected %d bytes, got %d", len(expected), n)
	}
}

func TestReaderReadeHeartbeatEvent(t *testing.T) {
	input := []byte{
		199, 205, 171, 241, 128, 63,
	}
	expected := []byte("49277 0")
	r := iotest.NewReadLogger(
		t.Name(),
		NewReader(bufio.NewReader(bytes.NewReader(input))),
	)
	result := make([]byte, len(expected))
	n, err := r.Read(result)
	if err != nil {
		t.Errorf("Error reading line: %s", err)
	}
	if !bytes.Equal(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
	if n != len(expected) {
		t.Errorf("Expected %d bytes, got %d", len(expected), n)
	}
}

func TestWriterWrite(t *testing.T) {
	input := []byte{
		102, 111, 111,
	}
	expected := []byte{
		3, 153, 144, 144, 255,
	}
	var result bytes.Buffer
	w := iotest.NewWriteLogger(
		t.Name(),
		NewWriter(bufio.NewWriter(&result)),
	)
	n, err := w.Write(input)
	if err != nil {
		t.Errorf("Error writing line: %s", err)
	}
	if !bytes.Equal(result.Bytes(), expected) {
		t.Errorf("Expected %v, got %v", expected, result.Bytes())
	}
	if n != len(input) {
		t.Errorf("Expected %d bytes, got %d", len(input), n)
	}
}
