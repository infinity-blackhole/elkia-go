package gateway

import (
	"bufio"
	"fmt"
	"io"

	"go.shikanime.studio/elkia/pkg/protonostale"
)

func ReadSession(r *bufio.Reader) ([]byte, error) {
	bs, err := r.ReadBytes(0x0E)
	if err != nil {
		return nil, err
	}
	buff := make([]byte, len(bs)*2)
	for n := 0; len(bs) > n; n++ {
		first_byte := bs[n] - 0xF
		second_byte := first_byte & 0xF0
		second_key := second_byte >> 0x4
		first_key := first_byte - second_byte
		for i, key := range []byte{second_key, first_key} {
			switch key {
			case 0, 1:
				buff[n*2+i] = ' '
			case 2:
				buff[n*2+i] = '-'
			case 3:
				buff[n*2+i] = '.'
			default:
				buff[n*2+i] = 0x2C + key
			}
		}
	}
	return buff, nil
}

type SessionDecoder struct {
	r *bufio.Reader
}

func NewSessionDecoder(r io.Reader) *SessionDecoder {
	return &SessionDecoder{bufio.NewReader(r)}
}

func (d *SessionDecoder) Decode(v any) error {
	buff, err := ReadSession(d.r)
	if err != nil {
		return err
	}
	if v, ok := v.(protonostale.Unmarshaler); ok {
		return v.UnmarshalNosTale(buff)
	}
	return fmt.Errorf("invalid payload: %v", v)
}
