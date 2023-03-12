package encoding

func DecodeFrame(dst, src []byte) (n int) {
	for n = 0; n < len(src); n++ {
		if src[n] > 14 {
			dst[n] = (src[n] - 15) ^ 195
		} else {
			dst[n] = (255 - (14 - src[n])) ^ 195
		}
	}
	return n
}

func EncodeFrame(dst, src []byte) (n int) {
	for n = 0; n < len(src); n++ {
		dst[n] = (src[n] + 15) & 0xFF
	}
	return n
}
