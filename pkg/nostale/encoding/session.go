package encoding

var SessionEncoding sessionEncoding

type sessionEncoding struct {
}

func (e sessionEncoding) Decode(dst, src []byte) (ndst, nsrc int, err error) {
	if len(dst) < len(src) {
		panic("dst buffer is too small")
	}
	for ndst, nsrc = 0, 0; len(src) > nsrc; ndst, nsrc = ndst+2, nsrc+1 {
		first_byte := src[nsrc] - 0xF
		second_byte := first_byte & 0xF0
		second_key := second_byte >> 0x4
		first_key := first_byte - second_byte
		for i, key := range []byte{second_key, first_key} {
			switch key {
			case 0, 1:
				dst[nsrc*2+i] = ' '
			case 2:
				dst[nsrc*2+i] = '-'
			case 3:
				dst[nsrc*2+i] = '.'
			default:
				dst[nsrc*2+i] = 0x2C + key
			}
		}
	}
	return ndst, nsrc, nil
}

func (e sessionEncoding) DecodedLen(x int) int {
	return x * 2
}

func (e sessionEncoding) Delim() byte {
	return 0xFF
}
