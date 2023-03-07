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
}

func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{
		r: bufio.NewReader(r),
	}
}

func (d *Decoder) Decode(v any) error {
	bs, err := d.r.ReadBytes('\n')
	if err != nil {
		return err
	}
	if err != nil {
		return err
	}
	switch v := v.(type) {
	case *string:
		*v = string(bs)
		return nil
	case *[]byte:
		*v = bs
		return nil
	case Unmarshaler:
		return v.UnmarshalNosTale(bs)
	default:
		return fmt.Errorf("invalid payload: %v", v)
	}
}

type Encoder struct {
	w *bufio.Writer
}

func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{
		w: bufio.NewWriter(w),
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
	if err != nil {
		return err
	}
	if _, err := e.w.Write(bs); err != nil {
		return err
	}
	if err := e.w.WriteByte('\n'); err != nil {
		return err
	}
	return e.w.Flush()
}
