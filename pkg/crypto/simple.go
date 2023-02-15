package crypto

import (
	"bufio"
)

// A Reader implements convenience methods for reading messages
// from a NosTale protocol network connection.
type ServerReader struct {
	R *bufio.Reader
}

// NewServerReader returns a new Reader reading from r.
//
// To avoid denial of service attacks, the provided bufio.Reader
// should be reading from an io.LimitReader or similar Reader to bound
// the size of responses.
func NewServerReader(r *bufio.Reader) *ServerReader {
	return &ServerReader{
		R: r,
	}
}

// ReadLine reads a single line from r,
// eliding the final \n or \r\n from the returned string.
func (r *ServerReader) ReadLine() (string, error) {
	line, err := r.readLineSlice()
	return string(line), err
}

// ReadLineBytes is like ReadLine but returns a []byte instead of a string.
func (r *ServerReader) ReadLineBytes() ([]byte, error) {
	line, err := r.readLineSlice()
	if line != nil {
		buf := make([]byte, len(line))
		copy(buf, line)
		line = buf
	}
	return line, err
}

func (r *ServerReader) readLineSlice() ([]byte, error) {
	line, err := r.R.ReadBytes(0xD8)
	if err != nil {
		return nil, err
	}
	buf := make([]byte, 0, len(line))
	for _, b := range line {
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
type ServerWriter struct {
	W *bufio.Writer
}

// NewWriter returns a new Writer writing to w.
func NewServerWriter(w *bufio.Writer) *ServerWriter {
	return &ServerWriter{
		W: w,
	}
}

func (r *ServerWriter) WriteLine(plaintext []byte) error {
	buf := make([]byte, 0, len(plaintext))
	for _, b := range plaintext {
		buf = append(buf, (b+15)&0xFF)
	}
	if _, err := r.W.Write(append(buf, 0x19, 0xD8)); err != nil {
		return err
	}
	return r.W.Flush()
}
