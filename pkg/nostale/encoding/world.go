package encoding

import (
	"bufio"
	"errors"
	"io"
)

type WorldReader struct {
	r      *bufio.Reader
	mode   byte
	offset byte
}

func NewWorldReader(r io.Reader, key uint32) *WorldReader {
	return &WorldReader{
		r:      bufio.NewReader(r),
		mode:   byte(key >> 6 & 0x03),
		offset: byte(key&0xFF + 0x40&0xFF),
	}

}

func (r *WorldReader) Read(p []byte) (n int, err error) {
	for n = 0; n < len(p); n++ {
		b, err := r.r.ReadByte()
		if err != nil {
			return n, err
		}
		switch r.mode {
		case 0:
			p[n] = b - r.offset
		case 1:
			p[n] = b + r.offset
		case 2:
			p[n] = (b - r.offset) ^ 0xC3
		case 3:
			p[n] = (b + r.offset) ^ 0xC3
		default:
			return n, errors.New("invalid mode")
		}
	}
	return n, nil
}

type WorldPackReader struct {
	r *bufio.Reader
}

func NewWorldPackReader(r io.Reader) *WorldPackReader {
	return &WorldPackReader{bufio.NewReader(r)}
}

func (r *WorldPackReader) Read(p []byte) (n int, err error) {
	bs, err := r.r.ReadBytes(0xFF)
	if err != nil {
		return 0, err
	}
	result := []byte{}
	for len(bs) > 0 {
		flag := bs[0]
		payload := bs[1:]
		if flag == 0xFF {
			bs = payload
			result = append(result, '\n')
		} else if flag <= 0x7A {
			first := make([]byte, len(payload))
			n := r.decodeLinearChunk(first, payload, flag)
			result = append(result, first[:n]...)
			bs = payload[n:]
		} else {
			first := make([]byte, len(payload)*2)
			ndst, nsrc := r.decodeCompactChunk(first, payload, flag&0x7F)
			bs = payload[nsrc:]
			result = append(result, first[:ndst]...)
		}
	}
	if len(result) > len(p) {
		panic("buffer too small")
	}
	return copy(p, result), nil
}

func (r *WorldPackReader) decodeLinearChunk(dst, src []byte, b byte) (n int) {
	var l int
	if int(b) < len(src) {
		l = int(b)
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

func (r *WorldPackReader) decodeCompactChunk(dst, src []byte, n byte) (ndst, nsrc int) {
	buff := src
	for ndst, nsrc = 0, 0; ndst < int(n) && len(buff) > 0; ndst, nsrc = ndst+1, nsrc+1 {
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

type WorldWriter struct {
	w *bufio.Writer
}

func NewWorldWriter(w io.Writer) *WorldWriter {
	return &WorldWriter{bufio.NewWriter(w)}
}

func (r *WorldWriter) Write(p []byte) (n int, err error) {
	var l int
	if p[len(p)-1] == '\n' {
		l = len(p) - 1
	} else {
		l = len(p)
	}
	for n := 0; n < len(p); n++ {
		if p[n] == '\n' {
			if err := r.w.WriteByte(0xFF); err != nil {
				return n, err
			}
		} else if (n % 0x7E) != 0 {
			if err := r.w.WriteByte(^p[n]); err != nil {
				return n, err
			}
		} else {
			remaining := byte(l - n)
			if remaining > 0x7E {
				remaining = 0x7E
			}
			if err := r.w.WriteByte(remaining); err != nil {
				return n, err
			}
			if err := r.w.WriteByte(^p[n]); err != nil {
				return n, err
			}
		}
	}
	return n, r.w.Flush()
}
