package nsl

func NewEncoding() *Encoding {
	return &Encoding{}
}

type Encoding struct {
}

func (e *Encoding) Decode(dst, src []byte) (n int, err error) {
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

func (e *Encoding) DecodedLen(x int) int {
	return x * 2
}

func (e *Encoding) Encode(dst, src []byte) (n int, err error) {
	if len(dst) < len(src) {
		panic("dst buffer is too small")
	}
	for n = 0; n < len(src); n++ {
		dst[n] = (src[n] + 15) & 0xFF
	}
	return n, nil
}

func (e *Encoding) EncodedLen(x int) int {
	return x
}

func (e *Encoding) Delim() byte {
	return 0xD8
}
