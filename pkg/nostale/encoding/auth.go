package encoding

import (
	"bufio"
	"io"
)

type AuthReader struct {
	r *bufio.Reader
}

func NewAuthReader(r io.Reader) *AuthReader {
	return &AuthReader{bufio.NewReader(r)}
}

func (e *AuthReader) Read(p []byte) (n int, err error) {
	for n = 0; n < len(p); n++ {
		b, err := e.r.ReadByte()
		if err != nil {
			return n, err
		}
		if b > 14 {
			p[n] = (b - 15) ^ 195
		} else {
			p[n] = (255 - (14 - b)) ^ 195
		}
		if p[n] == '\n' {
			return n + 1, nil
		}
	}
	return n, nil
}

type AuthWriter struct {
	w *bufio.Writer
}

func NewAuthWriter(w io.Writer) *AuthWriter {
	return &AuthWriter{bufio.NewWriter(w)}
}

func (e AuthWriter) Write(p []byte) (n int, err error) {
	for n = 0; n < len(p); n++ {
		if err := e.w.WriteByte((p[n] + 15) & 0xFF); err != nil {
			return n, err
		}
	}
	return n, e.w.Flush()
}
