package gateway

import (
	"bufio"
	"bytes"
	"errors"
	"math"
)

type Unmarshaler interface {
	UnmarshalNosTale([]byte) error
}

type Marshaler interface {
	MarshalNosTale() ([]byte, error)
}

type Encoding interface {
	Encode(dst, src []byte)
	Decode(dst, src []byte) (int, error)
	DecodedLen(x int) int
	Delim() byte
}

type SessionEncoding struct {
	r *bufio.Reader
}

func NewSessionDecoding() *SessionEncoding {
	return &SessionEncoding{}
}

func (e *SessionEncoding) Decode(dst, src []byte) (n int, err error) {
	if len(dst) < len(src) {
		panic("dst buffer is too small")
	}
	for n = 0; len(src) > n*2; n++ {
		first_byte := src[n] - 0xF
		second_byte := first_byte & 0xF0
		second_key := second_byte >> 0x4
		first_key := first_byte - second_byte
		for i, key := range []byte{second_key, first_key} {
			switch key {
			case 0, 1:
				dst[n*2+i] = ' '
			case 2:
				dst[n*2+i] = '-'
			case 3:
				dst[n*2+i] = '.'
			default:
				dst[n*2+i] = 0x2C + key
			}
		}
	}
	return n, nil
}

func (e *SessionEncoding) DecodedLen(x int) int {
	return x * 2
}

func (e *SessionEncoding) Encode(dst, src []byte) {
	if len(dst) < len(src) {
		panic("dst buffer is too small")
	}
	for n := 0; n < len(src); n++ {
		if (n % 0x7E) != 0 {
			dst[n] = ^src[n]
		} else {
			remaining := byte(len(src) - n)
			if remaining > 0x7E {
				remaining = 0x7E
			}
			dst[n] = remaining
			dst[n] = ^src[n]
		}
	}
}

func (e *SessionEncoding) Delim() byte {
	return '\n'
}

func NewWorldFrameListEncoding(key uint32) *FrameEncoding {
	return &FrameEncoding{
		mode:   byte(key >> 6 & 0x03),
		offset: byte(key&0xFF + 0x40&0xFF),
	}
}

type FrameEncoding struct {
	mode   byte
	offset byte
}

func (e *FrameEncoding) Decode(dst, src []byte) (n int, err error) {
	if len(dst) < len(src) {
		panic("dst buffer is too small")
	}
	n, err = e.decodeFrameList(dst, src)
	if err != nil {
		return n, err
	}
	n, err = e.unpackFrameList(dst, dst[:n])
	if err != nil {
		return n, err
	}
	return n, nil
}

func (e *FrameEncoding) decodeFrameList(dst, src []byte) (n int, err error) {
	for n = 0; n < len(src); n++ {
		switch e.mode {
		case 0:
			dst[n] = src[n] - e.offset
		case 1:
			dst[n] = src[n] + e.offset
		case 2:
			dst[n] = (src[n] - e.offset) ^ 0xC3
		case 3:
			dst[n] = (src[n] + e.offset) ^ 0xC3
		default:
			return n, errors.New("invalid mode")
		}
	}
	return n, nil
}

func (d *FrameEncoding) unpackFrameList(dst, src []byte) (n int, err error) {
	var lookup = []byte{
		'\x00', ' ', '-', '.', '0', '1', '2', '3', '4',
		'5', '6', '7', '8', '9', '\n', '\x00',
	}
	var sub [][]byte
	for _, s := range bytes.Split(src, []byte{0xFF}) {
		var result [][]byte
		for len(s) > 0 {
			head := s[0]
			s = s[1:]
			isPacked := head&0x80 > 0
			tmpLen := head & 0x7F
			var partLen int
			if isPacked {
				partLen = int(math.Ceil(float64(tmpLen) / 2))
			} else {
				partLen = int(tmpLen)
			}
			if partLen == 0 {
				continue
			}
			var chunk []byte
			if partLen <= len(lookup) {
				chunk = append(chunk, lookup[partLen-1])
			} else {
				chunk = make([]byte, partLen)
				for i := 0; i < partLen; i++ {
					chunk[i] = byte(0xFF)
				}
			}
			if isPacked {
				var decodedChunk []byte
				for i := 0; i < partLen/2; i++ {
					b := s[0]
					s = s[1:]
					hi := (b >> 4) & 0xF
					lo := b & 0xF
					decodedChunk = append(decodedChunk, chunk[hi], chunk[lo])
				}
				if partLen%2 == 1 {
					decodedChunk = append(decodedChunk, chunk[s[0]>>4])
				}
				chunk = decodedChunk
			}
			result = append(result, chunk)
		}
		sub = append(sub, bytes.Join(result, []byte{}))
	}
	copy(dst, bytes.Join(sub, []byte{0xFF}))
	return len(dst), nil
}

func (e *FrameEncoding) DecodedLen(x int) int {
	return x
}

func (e *FrameEncoding) Encode(dst, src []byte) {
	if len(dst) < len(src) {
		panic("dst buffer is too small")
	}
	for n := 0; n < len(src); n++ {
		if (n % 0x7E) != 0 {
			dst[n] = ^src[n]
		} else {
			remaining := byte(len(src) - n)
			if remaining > 0x7E {
				remaining = 0x7E
			}
			dst[n] = remaining
			dst[n] = ^src[n]
		}
	}
}

func (e *FrameEncoding) Delim() byte {
	return '\n'
}
