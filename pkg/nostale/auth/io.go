package auth

import (
	"bufio"
	"fmt"
	"io"
)

type Encoding interface {
	Encode(dst, src []byte)
	Decode(dst, src []byte) (int, error)
	DecodedLen(x int) int
	Delim() byte
}

type Decoder struct {
	r *bufio.Reader
	e Encoding
}

func NewDecoder(e Encoding, r io.Reader) *Decoder {
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
	buff := make([]byte, len(bs))
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
	e Encoding
}

func NewEncoder(e Encoding, w io.Writer) *Encoder {
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
	buff := make([]byte, len(bs))
	e.e.Encode(buff, bs)
	if _, err := e.w.Write(buff); err != nil {
		return err
	}
	if err := e.w.WriteByte(e.e.Delim()); err != nil {
		return err
	}
	return e.w.Flush()
}

type Unmarshaler interface {
	UnmarshalNosTale([]byte) error
}

type Marshaler interface {
	MarshalNosTale() ([]byte, error)
}

func NewLoginEncoding() *LoginEncoding {
	return &LoginEncoding{}
}

type LoginEncoding struct {
}

func (e *LoginEncoding) Decode(dst, src []byte) (n int, err error) {
	if len(dst) < len(src) {
		panic("dst buffer is too small")
	}
	if len(src) == 0 {
		return 0, nil
	}
	for n = 0; n < len(src); n++ {
		if src[n] > 14 {
			dst[n] = (src[n] - 15) ^ 195
		} else {
			dst[n] = (255 - (14 - src[n])) ^ 195
		}
	}
	return n, nil
}

func (e *LoginEncoding) DecodedLen(x int) int {
	return x * 2
}

func (e *LoginEncoding) Encode(dst, src []byte) {
	if len(dst) < len(src) {
		panic("dst buffer is too small")
	}
	for n := 0; n < len(src); n++ {
		dst[n] = (src[n] + 15) & 0xFF
	}
}

func (e *LoginEncoding) Delim() byte {
	return 0xD8
}
