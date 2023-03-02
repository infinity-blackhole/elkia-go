package encoding

import (
	"bufio"
	"fmt"
	"io"
)

type EncodingEncoder interface {
	Encode(dst, src []byte) (int, error)
	EncodedLen(x int) int
	Delim() byte
}

type EncodingDecoder interface {
	Decode(dst, src []byte) (int, error)
	DecodedLen(x int) int
	Delim() byte
}

type Unmarshaler interface {
	UnmarshalNosTale([]byte) error
}

type Marshaler interface {
	MarshalNosTale() ([]byte, error)
}

type Decoder struct {
	r *bufio.Reader
	e EncodingDecoder
}

func NewDecoder(e EncodingDecoder, r io.Reader) *Decoder {
	return &Decoder{
		r: bufio.NewReader(r),
		e: e,
	}
}

func (d *Decoder) More() bool {
	return d.r.Buffered() > 0
}

func (d *Decoder) Decode(v any) error {
	bs, err := d.r.ReadBytes(d.e.Delim())
	if err != nil {
		return err
	}
	buff := make([]byte, d.e.DecodedLen(len(bs)))
	if _, err := d.e.Decode(buff, bs); err != nil {
		return err
	}
	switch v := v.(type) {
	case *string:
		*v = string(buff)
		return nil
	case *[]byte:
		*v = buff
		return nil
	case Unmarshaler:
		return v.UnmarshalNosTale(buff)
	default:
		return fmt.Errorf("invalid payload: %v", v)
	}
}

type Encoder struct {
	w *bufio.Writer
	e EncodingEncoder
}

func NewEncoder(e EncodingEncoder, w io.Writer) *Encoder {
	return &Encoder{
		w: bufio.NewWriter(w),
		e: e,
	}
}

func (e *Encoder) Encode(v any) (err error) {
	var bs []byte
	switch v := v.(type) {
	case *string:
		bs = []byte(*v)
	case *[]byte:
		bs = *v
	case Marshaler:
		bs, err = v.MarshalNosTale()
		if err != nil {
			return err
		}
	}
	buff := make([]byte, e.e.EncodedLen(len(bs)))
	n, err := e.e.Encode(buff, bs)
	if err != nil {
		return err
	}
	if _, err := e.w.Write(buff[:n]); err != nil {
		return err
	}
	if err := e.w.WriteByte(e.e.Delim()); err != nil {
		return err
	}
	return e.w.Flush()
}
