package auth

import (
	"bufio"
	"fmt"
	"io"
)

const Delim = byte(0xD8)

type Decoder struct {
	r *bufio.Reader
}

func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{
		r: bufio.NewReader(r),
	}
}

func (d *Decoder) More() bool {
	return d.r.Buffered() > 0
}

func (d *Decoder) Decode(v any) error {
	bs, err := d.r.ReadBytes(Delim)
	if err != nil {
		return err
	}
	buff := make([]byte, len(bs))
	if _, err := DecodeAuthFrame(buff, bs); err != nil {
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
	buff := make([]byte, len(bs))
	if _, err := EncodeAuthFrame(buff, bs); err != nil {
		return err
	}
	if _, err := e.w.Write(buff); err != nil {
		return err
	}
	if err := e.w.WriteByte(Delim); err != nil {
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

func DecodeAuthFrame(dst, src []byte) (n int, err error) {
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

func EncodeAuthFrame(dst, src []byte) (n int, err error) {
	if len(dst) < len(src) {
		panic("dst buffer is too small")
	}
	if len(src) == 0 {
		return 0, nil
	}
	for n = 0; n < len(src); n++ {
		dst[n] = (src[n] + 15) & 0xFF
	}
	return n, nil
}
