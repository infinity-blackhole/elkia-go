package encoding

import (
	"bytes"
	"errors"
	"math"
)

var WorldEncoding worldEncoding

type worldEncoding struct {
	mode   byte
	offset byte
}

func (e worldEncoding) WithKey(key uint32) *worldEncoding {
	return &worldEncoding{
		mode:   byte(key >> 6 & 0x03),
		offset: byte(key&0xFF + 0x40&0xFF),
	}
}

func (e worldEncoding) Decode(dst, src []byte) (ndst, nsrc int, err error) {
	if len(dst) < len(src) {
		panic("dst buffer is too small")
	}
	ndst, err = e.decodeFrameList(dst, src)
	if err != nil {
		return ndst, nsrc, err
	}
	nsrc, err = e.unpackFrameList(dst, dst[:ndst])
	if err != nil {
		return ndst, nsrc, err
	}
	return ndst, nsrc, nil
}

func (e worldEncoding) decodeFrameList(dst, src []byte) (n int, err error) {
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

func (e worldEncoding) unpackFrameList(dst, src []byte) (n int, err error) {
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

func (e worldEncoding) DecodedLen(x int) int {
	return x
}

func (e worldEncoding) Encode(dst, src []byte) (ndst, nsrc int, err error) {
	if len(dst) < len(src) {
		panic("dst buffer is too small")
	}
	for ndst, nsrc = 0, 0; nsrc < len(src); ndst, nsrc = ndst+1, nsrc+1 {
		if (nsrc % 0x7E) != 0 {
			dst[ndst] = ^src[nsrc]
		} else {
			remaining := byte(len(src) - nsrc)
			if remaining > 0x7E {
				remaining = 0x7E
			}
			dst[ndst] = remaining
			ndst++
			dst[ndst] = ^src[nsrc]
		}
	}
	return ndst, nsrc, nil
}

func (e worldEncoding) EncodedLen(x int) int {
	return x * 2
}

func (e worldEncoding) Delim() byte {
	return '\n'
}
