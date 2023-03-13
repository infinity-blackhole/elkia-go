package gateway

import (
	"bufio"
	"bytes"
	"fmt"
	"io"

	"github.com/infinity-blackhole/elkia/pkg/protonostale"
)

type SessionScanner struct {
	s *bufio.Scanner
}

func NewSessionScanner(r io.Reader) *SessionScanner {
	s := bufio.NewScanner(r)
	s.Split(ScanSessionFrame)
	return &SessionScanner{s}
}

func (s *SessionScanner) Scan() bool {
	return s.s.Scan()
}

func (s *SessionScanner) Err() error {
	return s.s.Err()
}

func (s *SessionScanner) Bytes() []byte {
	bs := s.s.Bytes()
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
	return buff
}

func (s *SessionScanner) Text() string {
	return string(s.Bytes())
}

func ScanSessionFrame(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.IndexByte(data, 0x0A); i >= 0 {
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

type SessionDecoder struct {
	s *SessionScanner
}

func NewSessionDecoder(r io.Reader) *SessionDecoder {
	return &SessionDecoder{NewSessionScanner(r)}
}

func (d *SessionDecoder) Decode(v any) error {
	if !d.s.Scan() {
		if err := d.s.Err(); err != nil {
			return err
		}
		return io.EOF
	}
	if v, ok := v.(protonostale.Unmarshaler); ok {
		return v.UnmarshalNosTale(d.s.Bytes())
	}
	return fmt.Errorf("invalid payload: %v", v)
}
