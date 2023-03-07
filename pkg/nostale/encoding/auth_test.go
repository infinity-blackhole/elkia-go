package encoding

import (
	"bytes"
	"io"
	"testing"
)

var (
	_ io.Reader = (*AuthReader)(nil)
	_ io.Writer = (*AuthWriter)(nil)
)

func TestAuthEncodingDecodeFrame(t *testing.T) {
	input := []byte(
		"\x9c\xbb\x9f\x02\x05\x03\x05\xf2\xff\xff\x04\x05\xff\x06\x0a\xf2\xc0" +
			"\xb9\xaf\xbb\xb4\xbb\x0a\xff\x05\x02\x92\xbb\xc6\xb1\xbc\xba\xbb" +
			"\xbd\xb5\xfc\xaf\xbb\xbd\xf2\x01\x05\x01\xff\x01\x09\x05\x04\x90" +
			"\x0a\xff\x04\x03\x09\x05\x04\xff\x00\x06\x03\xff\x03\x01\x04\x05" +
			"\x09\xff\x03\x06\x03\x06\x04\x05\x09\x00\x06\x05\x03\x06\xff\x90" +
			"\x00\x01\x04\x96\x05\x00\xff\x94\x04\x05\x06\x0a\x95\x00\x03\x90" +
			"\x00\xf2\x02\x02\x04\x94\x03\x06\x06\x04\xd7\x02\xfc\x09\xfc\xff" +
			"\xfc\xff\x02\x0a\x04\xd8",
	)
	expected := []byte(
		"NoS0575 3365348 ricofo8350@otanhome.com 15131956B8367956324737165937" +
			"474659245743B216D523F6548E27B2 006F7446\v0.9.3.3086\n",
	)
	result, err := io.ReadAll(NewAuthReader(bytes.NewReader(input)))
	if err != nil {
		t.Errorf("Error reading line: %s", err)
	}
	if !bytes.Equal(expected, result) {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

func TestAuthEncodingEncodeFrame(t *testing.T) {
	input := []byte("fail Hello. This is a basic test\n")
	expected := []byte(
		"\x75\x70\x78\x7b\x2f\x57\x74\x7b\x7b\x7e\x3d\x2f\x63\x77\x78\x82\x2f" +
			"\x78\x82\x2f\x70\x2f\x71\x70\x82\x78\x72\x2f\x83\x74\x82\x83\x19",
	)
	var buf bytes.Buffer
	n, err := NewAuthWriter(&buf).Write(input)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	result := buf.Bytes()
	if !bytes.Equal(expected, result[:n]) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}
