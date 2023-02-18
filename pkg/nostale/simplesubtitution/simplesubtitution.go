package simplesubtitution

import (
	"bufio"
)

// A Reader implements convenience methods for reading messages
// from a NosTale protocol network connection.
type Reader struct {
	R *bufio.Reader
}

// NewReader returns a new Reader reading from r.
//
// To avoid denial of service attacks, the provided bufio.Reader
// should be reading from an io.LimitReader or similar Reader to bound
// the size of responses.
func NewReader(r *bufio.Reader) *Reader {
	return &Reader{
		R: r,
	}
}

// ReadMessage reads a single message from r,
// eliding the final \n or \r\n from the returned string.
func (r *Reader) ReadMessage() (string, error) {
	msg, err := r.readMessageSlice()
	return string(msg), err
}

// ReadMessageBytes is like ReadMessage but returns a []byte instead of a
// string.
func (r *Reader) ReadMessageBytes() ([]byte, error) {
	msg, err := r.readMessageSlice()
	if msg != nil {
		buf := make([]byte, len(msg))
		copy(buf, msg)
		msg = buf
	}
	return msg, err
}

func (r *Reader) readMessageSlice() ([]byte, error) {
	msg, err := r.R.ReadBytes(0xD8)
	if err != nil {
		return nil, err
	}
	buf := make([]byte, 0, len(msg))
	for _, b := range msg {
		if b > 14 {
			buf = append(buf, (b-15)^195)
		} else {
			buf = append(buf, (255-(14-b))^195)
		}
	}
	return buf, nil
}

// A Writer implements convenience methods for writing
// messages to a NosTale protocol network connection.
type Writer struct {
	W *bufio.Writer
}

// NewWriter returns a new Writer writing to w.
func NewWriter(w *bufio.Writer) *Writer {
	return &Writer{
		W: w,
	}
}

// WriteMessage writes the formatted message.
func (w *Writer) WriteMessage(msg []byte) error {
	buf := make([]byte, 0, len(msg))
	for _, b := range msg {
		buf = append(buf, (b+15)&0xFF)
	}
	if _, err := w.W.Write(append(buf, 0x19, 0xD8)); err != nil {
		return err
	}
	return w.W.Flush()
}
