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

func (r *Reader) Read(p []byte) (n int, err error) {
	for n = 0; n < len(p); n++ {
		// read the next byte
		c, err := r.r.ReadByte()
		if err != nil {
			return n, err
		}
		// write the decrypted byte to the output
		if c > 14 {
			p[n] = (c - 15) ^ 195
		} else {
			p[n] = (255 - (14 - c)) ^ 195
		}
		// if this is the end of a message, return the bytes read so far
		if c == 0xD8 {
			return n + 1, nil
		}
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
	if err := w.w.WriteByte(0x19); err != nil {
		return n, err
	}
	if err := w.w.WriteByte(0xD8); err != nil {
		return n, err
	}
	return n, w.w.Flush()
}
