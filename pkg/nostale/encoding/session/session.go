package encoding

func DecodeFrame(dst, src []byte) (n int) {
	for n = 0; len(src) > n; n++ {
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
	return n * 2
}

func DecodeFrameMaxLen(x int) int {
	return x * 2
}
