package encoding

import (
	"bufio"
	"fmt"
	"io"
)

type EncodingEncoder interface {
	Encode(dst, src []byte) (ndst, nsrc int, err error)
	EncodedLen(x int) int
	Delim() byte
}

type EncodingDecoder interface {
	Decode(dst, src []byte) (ndst, nsrc int, err error)
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

func NewDecoder(r io.Reader, e EncodingDecoder) *Decoder {
	return &Decoder{
		r: bufio.NewReader(r),
		e: e,
	}
}

func (d *Decoder) Decode(v any) error {
	bs, err := d.r.ReadBytes(d.e.Delim())
	if err != nil {
		return err
	}
	buff := make([]byte, d.e.DecodedLen(len(bs)))
	ndst, _, err := d.e.Decode(buff, bs)
	if err != nil {
		return err
	}
	switch v := v.(type) {
	case *string:
		*v = string(buff[:ndst])
		return nil
	case *[]byte:
		*v = buff[:ndst]
		return nil
	case Unmarshaler:
		return v.UnmarshalNosTale(buff[:ndst])
	default:
		return fmt.Errorf("invalid payload: %v", v)
	}
}

type Encoder struct {
	w *bufio.Writer
	e EncodingEncoder
}

func NewEncoder(w io.Writer, e EncodingEncoder) *Encoder {
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
	ndst, _, err := e.e.Encode(buff, bs)
	if err != nil {
		return err
	}
	if _, err := e.w.Write(buff[:ndst]); err != nil {
		return err
	}
	if err := e.w.WriteByte(e.e.Delim()); err != nil {
		return err
	}
	return e.w.Flush()
}
