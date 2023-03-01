package protonostale

import (
	"bytes"
	"testing"
)

// key:3324008792

func TestEncodeAuthFrame(t *testing.T) {
	input := []byte("fail Hello. This is a basic test")
	expected := []byte{
		117, 112, 120, 123, 47, 87, 116, 123, 123, 126, 61, 47, 99, 119, 120,
		130, 47, 120, 130, 47, 112, 47, 113, 112, 130, 120, 114, 47, 131, 116,
		130, 131,
	}
	result := make([]byte, len(input))
	n, err := EncodeAuthFrame(result, input)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	if !bytes.Equal(expected, result[:n]) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestDecodeAuthFrame(t *testing.T) {
	input := []byte{
		156, 187, 159, 2, 5, 3, 5, 242, 255, 4, 1, 6, 2, 255, 10, 242, 177,
		242, 5, 145, 149, 4, 0, 5, 4, 4, 5, 148, 255, 149, 2, 144, 150, 2, 145,
		2, 4, 5, 149, 150, 2, 3, 145, 6, 1, 9, 10, 9, 149, 6, 2, 0, 5, 144, 3,
		9, 150, 1, 255, 9, 255, 2, 145, 0, 145, 10, 143, 5, 3, 150, 4, 144, 6,
		255, 0, 5, 0, 0, 4, 3, 2, 3, 150, 9, 5, 4, 145, 2, 10, 0, 150, 1, 149,
		9, 1, 144, 6, 150, 9, 4, 145, 3, 9, 255, 5, 4, 0, 150, 148, 9, 10,
		148, 150, 2, 255, 143, 9, 150, 143, 148, 3, 6, 255, 143, 9, 143, 3,
		144, 6, 149, 255, 2, 5, 5, 150, 6, 148, 9, 148, 2, 9, 144, 145, 2, 1,
		5, 242, 2, 2, 255, 9, 149, 255, 150, 143, 215, 2, 252, 9, 252, 255,
		252, 255, 2, 3, 1, 242, 2, 242, 143, 3, 150, 0, 5, 2, 255, 144, 150,
		0, 5, 3, 148, 5, 144, 145, 149, 2, 10, 3, 2, 148, 6, 2, 143, 0, 150,
		145, 255, 4, 4, 4, 216,
	}
	expected := []byte(
		"NoS0575 3614038 a 5AE625665F3E0BD0A065ED07A41989E4025B79D139" +
			"30A2A8C57D6B4325226707D956A082D1E91B4D96A793562DF98FD03C9DCF743" +
			"C9C7B4E3055D4F9F09BA015 0039E3DC\v0.9.3.3071 0 C7D2503BD257F5" +
			"BAE0870F40C2DA3666\n",
	)
	result := make([]byte, len(input))
	n, err := DecodeAuthFrame(result, input)
	if err != nil {
		t.Errorf("Error reading line: %s", err)
	}
	if !bytes.Equal(expected, result[:n]) {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

func TestEncodeDecodeAuthFrame(t *testing.T) {
	input := []byte("fail Hello. This is a basic test")
	buff := make([]byte, len(input))
	n, err := EncodeAuthFrame(buff, input)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	result := make([]byte, len(input))
	n, err = DecodeAuthFrame(result, buff[:n])
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	if bytes.Equal(input, result[:n]) {
		t.Errorf("Expected different from %s, got %s", input, result)
	}
}

func TestReaderReadSyncEvent(t *testing.T) {
	input := []byte{
		150, 156, 122, 80, 79, 14, 198, 205, 171, 145, 70, 205, 214, 220, 208,
		217, 208, 196, 7, 212, 73,
	}
	expected := []byte("4349270 0 ;;737:584-.37:83898 868 71;481.6")
	result := make([]byte, DecodedSessionFrameLen(len(input)))
	n, err := DecodeSessionFrame(result, input)
	if err != nil {
		t.Errorf("Error reading line: %s", err)
	}
	if !bytes.Equal(expected, result[:n]) {
		t.Errorf("Expected %v, got %v", expected, result[:n])
	}
}

func TestReaderReadAuthHandoffPasswordEvent(t *testing.T) {
	input := []byte{
		208, 203, 222, 209, 215, 208, 210, 218, 193, 112, 67, 220, 208, 210,
		63, 199, 205, 171, 161, 16, 72, 215, 214, 221, 200, 214, 200, 214, 248,
		193, 160, 65, 218, 193, 224, 66, 241, 205, 63,
	}
	expected := []byte("49272 9hibwiwiG2e6Nr")
	frameList := make([]byte, len(input))
	n, err := DecodeFrameList(frameList, input, 0)
	if err != nil {
		t.Errorf("Error reading line: %s", err)
	}
	frames := bytes.Split(frameList, []byte{0xFF})
	pass := make([]byte, len(frames[0]))
	n, err = DecodeFrame(pass, frames[0])
	if err != nil {
		t.Errorf("Error reading line: %s", err)
	}
	if !bytes.Equal(frameList, expected) {
		t.Errorf("Expected %v, got %v", expected, frameList)
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
	frameList := make([]byte, len(expected))
	n, err := DecodeFrameList(frameList, input, 0)
	if err != nil {
		t.Errorf("Error reading line: %s", err)
	}
	frames := bytes.Split(frameList, []byte{0xFF})
	result := make([]byte, len(expected))
	n, err = DecodeFrame(result, frames[0])
	if err != nil {
		t.Errorf("Error reading line: %s", err)
	}
	if !bytes.Equal(expected, result[:n]) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestWriterWrite(t *testing.T) {
	input := []byte("foo")
	expected := []byte{
		3, 153, 144, 144,
	}
	result := make([]byte, MaxEncodedFrameLen(len(input)))
	n, _, err := EncodeFrame(result, input)
	if err != nil {
		t.Errorf("Error writing line: %s", err)
	}
	if !bytes.Equal(expected, result[:n]) {
		t.Errorf("Expected %v, got %v", expected, result[:n])
	}
}
