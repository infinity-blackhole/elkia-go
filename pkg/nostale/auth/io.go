package auth

import (
	"bufio"
)

// NewReader returns a new Reader reading from r.
//
// To avoid denial of service attacks, the provided bufio.Reader
// should be reading from an io.LimitReader or similar Reader to bound
// the size of responses.
func NewReader(r *bufio.Reader) *Reader {
	return &Reader{
		r: r,
	}
}

// A Reader implements convenience methods for reading messages
// from a NosTale protocol network connection.
type Reader struct {
	r *bufio.Reader
}

func (r *Reader) ReadByte() (byte, error) {
	c, err := r.r.ReadByte()
	if err != nil {
		return c, err
	}
	if c > 14 {
		return (c - 15) ^ 195, nil
	}
	return (255 - (14 - c)) ^ 195, nil
}

func (r *Reader) Read(p []byte) (n int, err error) {
	for n = 0; n < len(p); n++ {
		c, err := r.ReadByte()
		if err != nil {
			return n, err
		}
		p[n] = c
	}
	return n, nil
}

// A Writer implements convenience methods for writing
// messages to a NosTale protocol network connection.
type Writer struct {
	w *bufio.Writer
}

// NewWriter returns a new Writer writing to w.
func NewWriter(w *bufio.Writer) *Writer {
	return &Writer{
		w: w,
	}
}

// Write writes the formatted message.
func (w *Writer) Write(p []byte) (n int, err error) {
	for n = 0; n < len(p); n++ {
		if err := w.w.WriteByte((p[n] + 15) & 0xFF); err != nil {
			return n, err
		}
	}
	return n, w.w.Flush()
}
