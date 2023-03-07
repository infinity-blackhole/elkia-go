package encoding

import (
	"bufio"
	"io"
)

type SessionReader struct {
	r *bufio.Reader
}

func NewSessionReader(r io.Reader) *SessionReader {
	return &SessionReader{bufio.NewReader(r)}
}

func (e SessionReader) Read(p []byte) (n int, err error) {
	for n = 0; len(p) > n; n++ {
		b, err := e.r.ReadByte()
		if err != nil {
			return n * 2, err
		}
		first_byte := b - 0xF
		second_byte := first_byte & 0xF0
		second_key := second_byte >> 0x4
		first_key := first_byte - second_byte
		for i, key := range []byte{second_key, first_key} {
			switch key {
			case 0, 1:
				p[n*2+i] = ' '
			case 2:
				p[n*2+i] = '-'
			case 3:
				p[n*2+i] = '.'
			default:
				p[n*2+i] = 0x2C + key
			}
		}
		if p[n*2] == '\r' && p[n*2+1] == '\n' {
			return n*2 + 2, nil
		}
	}
	return n * 2, nil
}
