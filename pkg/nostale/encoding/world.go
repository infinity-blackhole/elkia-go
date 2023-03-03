package encoding

import (
	"bytes"
	"errors"
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
	var chunks [][]byte
	for _, chunk := range bytes.Split(src, []byte{0xFF}) {
		chunks = append(chunks, doDecryptHelper(chunk, [][]byte{}))
	}
	result := bytes.Join(chunks, []byte{})
	copy(dst, result)
	return len(result), nil
}

func doDecryptHelper(binary []byte, result [][]byte) []byte {
	if len(binary) == 0 {
		return []byte(reverseAndJoin(result))
	}

	b := binary[0]
	rest := binary[1:]

	if b <= 0x7A {
		var l int
		if int(b) < len(rest) {
			l = int(b)
		} else {
			l = len(rest)
		}

		first := rest[:l]
		second := rest[l:]

		res := make([]byte, len(first))
		for _, c := range first {
			res = append(res, c^0xFF)
		}

		return doDecryptHelper(second, append([][]byte{res}, result...))
	} else {
		first, second := doDecrypt2(rest, b&0x7F)
		return doDecryptHelper(second, append([][]byte{first}, result...))
	}
}

func doDecrypt2(binary []byte, n byte) ([]byte, []byte) {
	i := 0
	result := []byte{}

	for i < int(n) && len(binary) > 0 {
		h := int(binary[0] >> 4)
		l := int(binary[0] & 0x0F)
		binary = binary[1:]

		if h != 0 && h != 0xF && (l == 0 || l == 0xF) {
			result = append(result, table[h-1])
		} else if l != 0 && l != 0xF && (h == 0 || h == 0xF) {
			result = append(result, table[l-1])
		} else if h != 0 && h != 0xF && l != 0 && l != 0xF {
			result = append(result, table[h-1], table[l-1])
		}
		i += 1
	}

	return result, binary
}

func reverseAndJoin(strings [][]byte) []byte {
	reversed := [][]byte{}
	for i := len(strings) - 1; i >= 0; i-- {
		reversed = append(reversed, strings[i])
	}
	return bytes.Join(reversed, []byte{})
}

var table = []byte{
	' ', '-', '.', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', 'n',
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
