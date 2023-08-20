package auth

import (
	"bufio"
	"bytes"
	"io"

	"go.shikanime.studio/elkia/pkg/protonostale"
)

type Scanner struct {
	s *bufio.Scanner
}

func NewScanner(r io.Reader) *Scanner {
	s := bufio.NewScanner(r)
	s.Split(ScanFrame)
	return &Scanner{s}
}

func (s *Scanner) Scan() bool {
	return s.s.Scan()
}

func (s *Scanner) Err() error {
	return s.s.Err()
}

func (s *Scanner) Bytes() []byte {
	bs := s.s.Bytes()
	for n := 0; n < len(bs); n++ {
		if bs[n] > 14 {
			bs[n] = (bs[n] - 15) ^ 195
		} else {
			bs[n] = (255 - (14 - bs[n])) ^ 195
		}
	}
	return bs
}

func (s *Scanner) Text() string {
	return string(s.Bytes())
}

func ScanFrame(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.IndexByte(data, 0xD8); i >= 0 {
		// We have a full frames.
		return i + 1, data[0:i], nil
	}
	// If we're at EOF, we have a final, non-terminated frame. Return it.
	if atEOF {
		return len(data), data, nil
	}
	// Request more data.
	return 0, nil, nil
}

type Decoder struct {
	s *Scanner
}

func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{NewScanner(r)}
}

func (d *Decoder) Decode(v any) error {
	if !d.s.Scan() {
		if err := d.s.Err(); err != nil {
			return err
		}
		return io.EOF
	}
	return protonostale.UnmarshalNosTale(d.s.Bytes(), v)
}

type Writer struct {
	w *bufio.Writer
}

func NewWriter(r io.Writer) *Writer {
	return &Writer{bufio.NewWriter(r)}
}

func (w *Writer) WriteFrame(b []byte) error {
	if err := w.Write(b); err != nil {
		return err
	}
	if err := w.w.WriteByte(0x19); err != nil {
		return err
	}
	return w.w.Flush()
}

func (w *Writer) Write(b []byte) error {
	for n := 0; n < len(b); n++ {
		if err := w.w.WriteByte((b[n] + 15) & 0xFF); err != nil {
			return err
		}
	}
	return nil
}

type Encoder struct {
	w *Writer
}

func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{NewWriter(w)}
}

func (e *Encoder) Encode(v any) (err error) {
	bs, err := protonostale.MarshalNosTale(v)
	if err != nil {
		return err
	}
	return e.w.WriteFrame(bs)
}
