package gateway

import (
	"bufio"
	"bytes"
	"fmt"
	"io"

	"go.shikanime.studio/elkia/pkg/protonostale"
)

type PackedChannelScanner struct {
	s            *bufio.Scanner
	mode, offset byte
}

func NewPackedChannelScanner(r io.Reader, key uint32) *PackedChannelScanner {
	s := bufio.NewScanner(r)
	s.Split(ScanCommandFrame)
	return &PackedChannelScanner{
		s:      s,
		mode:   byte(key >> 6 & 0x03),
		offset: byte(key&0xFF + 0x40&0xFF),
	}
}

func (s *PackedChannelScanner) Scan() bool {
	return s.s.Scan()
}

func (s *PackedChannelScanner) Err() error {
	return s.s.Err()
}

func (s *PackedChannelScanner) Bytes() []byte {
	bs := s.s.Bytes()
	result := make([]byte, len(bs))
	for n := 0; n < len(bs); n++ {
		switch s.mode {
		case 0:
			result[n] = bs[n] - s.offset
		case 1:
			result[n] = bs[n] + s.offset
		case 2:
			result[n] = (bs[n] - s.offset) ^ 0xC3
		case 3:
			result[n] = (bs[n] + s.offset) ^ 0xC3
		}
	}
	return result
}

func (s *PackedChannelScanner) Text() string {
	return string(s.Bytes())
}

func ScanCommandFrame(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.IndexByte(data, 0x3F); i >= 0 {
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

type ChannelScanner struct {
	s *PackedChannelScanner
}

func NewChannelScanner(r io.Reader, key uint32) *ChannelScanner {
	return &ChannelScanner{NewPackedChannelScanner(r, key)}
}

func (s *ChannelScanner) Scan() bool {
	return s.s.Scan()
}

func (s *ChannelScanner) Err() error {
	return s.s.Err()
}

func (s *ChannelScanner) Bytes() []byte {
	bs := s.s.Bytes()
	result := []byte{}
	for len(bs) > 0 {
		flag := bs[0]
		payload := bs[1:]
		if flag <= 0x7A {
			first := make([]byte, len(payload))
			n := s.decodePackedLinearFrame(first, payload, flag)
			result = append(result, first[:n]...)
			bs = payload[n:]
		} else {
			first := make([]byte, len(payload)*2)
			ndst, nsrc := s.decodePackedCompactFrame(first, payload, flag&0x7F)
			bs = payload[nsrc:]
			result = append(result, first[:ndst]...)
		}
	}
	return result
}

func (s *ChannelScanner) decodePackedLinearFrame(dst, src []byte, flag byte) (n int) {
	var l int
	if int(flag) < len(src) {
		l = int(flag)
	} else {
		l = len(src)
	}
	for n = 0; n < l; n++ {
		dst[n] = src[n] ^ 0xFF
	}
	return n
}

var permutations = []byte{
	' ', '-', '.', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', 'n',
}

func (s *ChannelScanner) decodePackedCompactFrame(dst, src []byte, flag byte) (ndst, nsrc int) {
	buff := src
	for ndst, nsrc = 0, 0; ndst < int(flag) && len(buff) > 0; ndst, nsrc = ndst+1, nsrc+1 {
		h := int(buff[0] >> 4)
		l := int(buff[0] & 0x0F)
		buff = buff[1:]
		if h != 0 && h != 0xF && (l == 0 || l == 0xF) {
			dst[ndst] = permutations[h-1]
		} else if l != 0 && l != 0xF && (h == 0 || h == 0xF) {
			dst[ndst] = permutations[l-1]
		} else if h != 0 && h != 0xF && l != 0 && l != 0xF {
			dst[ndst] = permutations[h-1]
			ndst++
			dst[ndst] = permutations[l-1]
		}
	}
	return ndst, nsrc
}

func (s *ChannelScanner) Text() string {
	return string(s.Bytes())
}

type ChannelDecoder struct {
	s *ChannelScanner
}

func NewChannelDecoder(r io.Reader, key uint32) *ChannelDecoder {
	return &ChannelDecoder{NewChannelScanner(r, key)}
}

func (d *ChannelDecoder) Decode(v any) error {
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
