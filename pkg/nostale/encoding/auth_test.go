package encoding

import (
	"bytes"
	"testing"
)

func TestAuthEncodingDecodeFrame(t *testing.T) {
	var input bytes.Buffer
	input.Write([]byte{
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
	})
	expected := []byte(
		"NoS0575 3614038 a 5AE625665F3E0BD0A065ED07A41989E4025B79D139" +
			"30A2A8C57D6B4325226707D956A082D1E91B4D96A793562DF98FD03C9DCF743" +
			"C9C7B4E3055D4F9F09BA015 0039E3DC\v0.9.3.3071 0 C7D2503BD257F5" +
			"BAE0870F40C2DA3666\n",
	)
	result := make([]byte, AuthEncoding.DecodedLen(len(input.Bytes())))
	if err := NewDecoder(&input, AuthEncoding).Decode(&result); err != nil {
		t.Errorf("Error reading line: %s", err)
	}
	if !bytes.Equal(result, expected) {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

func TestAuthEncodingEncodeFrame(t *testing.T) {
	input := "fail Hello. This is a basic test"
	expected := []byte{
		117, 112, 120, 123, 47, 87, 116, 123, 123, 126, 61, 47, 99, 119, 120,
		130, 47, 120, 130, 47, 112, 47, 113, 112, 130, 120, 114, 47, 131, 116,
		130, 131, 216,
	}
	var buf bytes.Buffer
	if err := NewEncoder(&buf, AuthEncoding).Encode(&input); err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	result := buf.Bytes()
	if !bytes.Equal(expected, result) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}
