package protonostale

import (
	"bytes"
	"errors"
	"math"
)

func DecodeAuthFrame(dst, src []byte) (n int, err error) {
	for n = 0; len(src) > n; n++ {
		if src[n] > 14 {
			dst[n] = (src[n] - 15) ^ 195
		} else {
			dst[n] = (255 - (14 - src[n])) ^ 195

		}
	}
	return n, nil
}

func EncodeAuthFrame(dst, src []byte) (n int, err error) {
	for n = 0; len(src) > n; n++ {
		dst[n] = (src[n] + 15) & 0xFF
	}
	return n, nil
}

func DecodeSessionFrame(dst, src []byte) (n int, err error) {
	for n = 0; len(src) > n; n++ {
		first_byte := src[n] - 0xF
		second_byte := first_byte & 0xF0
		second_key := second_byte >> 0x4
		first_key := first_byte - second_byte
		for i, b := range []byte{second_key, first_key} {
			switch b {
			case 0, 1:
				dst[n*2+i] = ' '
			case 2:
				dst[n*2+i] = '-'
			case 3:
				dst[n*2+i] = '.'
			default:
				dst[n*2+i] = 0x2C + b
			}
		}
	}
	return n * 2, nil
}

func DecodedSessionFrameLen(x int) int {
	return x * 2
}

func DecodeFrameList(dst, src []byte, key uint32) (n int, err error) {
	mode := byte(key >> 6 & 0x03)
	offset := byte(key&0xFF + 0x40&0xFF)
	for n = 0; len(src) > n; n++ {
		switch mode {
		case 0:
			src[n] = src[n] - offset
		case 1:
			src[n] = src[n] + offset
		case 2:
			src[n] = (src[n] - offset) ^ 0xC3
		case 3:
			src[n] = (src[n] + offset) ^ 0xC3
		default:
			return n, errors.New("invalid mode")
		}
	}
	return n, nil
}

func DecodeFrame(dst, src []byte) (int, error) {
	var lookup = []byte{
		'\x00', ' ', '-', '.', '0', '1', '2', '3', '4',
		'5', '6', '7', '8', '9', '\n', '\x00',
	}
	var chunks [][]byte
	for len(src) > 0 {
		head := src[0]
		src = src[1:]
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
				b := src[0]
				src = src[1:]
				hi := (b >> 4) & 0xF
				lo := b & 0xF
				decodedChunk = append(decodedChunk, chunk[hi], chunk[lo])
			}
			if partLen%2 == 1 {
				decodedChunk = append(decodedChunk, chunk[src[0]>>4])
			}
			chunk = decodedChunk
		}
		chunks = append(chunks, chunk)
	}
	copy(dst, bytes.Join(chunks, []byte{}))
	return len(dst), nil
}

func EncodeFrame(dst, src []byte) (ndst, nsrc int, err error) {
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

func MaxEncodedFrameLen(n int) int {
	return n * 2
}
