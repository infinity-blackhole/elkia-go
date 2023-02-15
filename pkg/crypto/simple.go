package crypto

import (
	"bufio"
	"io"
)

type ServerReader struct {
	r *bufio.Reader
}

func NewServerReader(r *bufio.Reader) *ServerReader {
	return &ServerReader{
		r: r,
	}
}

func (r *ServerReader) ReadLineBytes() ([]byte, error) {
	ciphertext, err := r.r.ReadBytes(0xD8)
	if err != nil {
		return nil, err
	}
	buf := make([]byte, 0, len(ciphertext))
	for _, b := range ciphertext {
		if b > 14 {
			buf = append(buf, (b-15)^195)
		} else {
			buf = append(buf, (255-(14-b))^195)
		}
	}
	return buf, nil
}

func (r *ServerReader) ReadLine() (string, error) {
	buf, err := r.ReadLineBytes()
	if err != nil {
		return "", err
	}
	return string(buf), nil
}

type ServerWriter struct {
	w io.Writer
}

func NewServerWriter(w io.Writer) *ServerWriter {
	return &ServerWriter{
		w: w,
	}
}

func (r *ServerWriter) Write(plaintext []byte) (int, error) {
	buf := make([]byte, 0, len(plaintext))
	for _, b := range plaintext {
		buf = append(buf, (b+15)&0xFF)
	}
	return r.w.Write(append(buf, 0x19, 0xD8))
}
