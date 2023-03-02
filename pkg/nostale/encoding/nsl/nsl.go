package nsl

func NewEncoding() *Encoding {
	return &Encoding{}
}

type Encoding struct {
}

func (e *Encoding) Decode(dst, src []byte) (ndst, nsrc int, err error) {
	if len(dst) < len(src) {
		panic("dst buffer is too small")
	}
	if len(src) == 0 {
		return ndst, nsrc, nil
	}
	for ndst, nsrc = 0, 0; nsrc < len(src); ndst, nsrc = ndst+1, nsrc+1 {
		if src[nsrc] > 14 {
			dst[ndst] = (src[nsrc] - 15) ^ 195
		} else {
			dst[ndst] = (255 - (14 - src[nsrc])) ^ 195
		}
	}
	return ndst, nsrc, nil
}

func (e *Encoding) DecodedLen(x int) int {
	return x
}

func (e *Encoding) Encode(dst, src []byte) (ndst, nsrc int, err error) {
	if len(dst) < len(src) {
		panic("dst buffer is too small")
	}
	for ndst, nsrc = 0, 0; nsrc < len(src); ndst, nsrc = ndst+1, nsrc+1 {
		dst[ndst] = (src[nsrc] + 15) & 0xFF
	}
	return ndst, nsrc, nil
}

func (e *Encoding) EncodedLen(x int) int {
	return x
}

func (e *Encoding) Delim() byte {
	return 0xD8
}
