package gateway

import (
	"bufio"
	"io"

	"github.com/infinity-blackhole/elkia/pkg/protonostale"
)

type Writer struct {
	w *bufio.Writer
}

func NewWriter(r io.Writer) *Writer {
	return &Writer{bufio.NewWriter(r)}
}

func (w *Writer) Write(b []byte) (n int, err error) {
	for i := 0; i < len(b); n, i = n+1, i+1 {
		if (i % 0x7E) != 0 {
			if err := w.w.WriteByte(^b[i]); err != nil {
				return n, err
			}
		} else {
			remaining := byte(len(b))
			if remaining > 0x7E {
				remaining = 0x7E
			}
			if err := w.w.WriteByte(remaining); err != nil {
				return n, err
			}
			n++
			if err := w.w.WriteByte(^b[i]); err != nil {
				return n, err
			}
		}
	}
	return n, err
}

func (w *Writer) WriteFrame(b []byte) error {
	if _, err := w.Write(b); err != nil {
		return err
	}
	if err := w.w.WriteByte(0x19); err != nil {
		return err
	}
	return w.w.Flush()
}

type Encoder struct {
	w *Writer
}

func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{NewWriter(w)}
}

func (e *Encoder) Encode(v any) (err error) {
	var bs []byte
	switch v := v.(type) {
	case protonostale.Marshaler:
		bs, err = v.MarshalNosTale()
		if err != nil {
			return err
		}
	}
	return e.w.WriteFrame(bs)
}
