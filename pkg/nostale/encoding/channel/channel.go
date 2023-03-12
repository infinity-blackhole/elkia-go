package encoding

func DecodeFrame(dst, src []byte, key byte) (n int) {
	mode, offset := DecodeKey(key)
	for n = 0; n < len(src); n++ {
		switch mode {
		case 0:
			dst[n] = src[n] - offset
		case 1:
			dst[n] = src[n] + offset
		case 2:
			dst[n] = (src[n] - offset) ^ 0xC3
		case 3:
			dst[n] = (src[n] + offset) ^ 0xC3
		}
	}
	return DecodePackedFrame(dst, dst[:n])
}

func DecodeKey(key byte) (byte, byte) {
	mode := byte(key >> 6 & 0x03)
	offset := byte(key&0xFF + 0x40&0xFF)
	return mode, offset
}

func DecodePackedFrame(dst, src []byte) (n int) {
	buff := []byte{}
	for len(src) > 0 {
		flag := src[0]
		payload := src[1:]
		if flag == 0xFF {
			src = payload
			buff = append(buff, '\n')
		} else if flag <= 0x7A {
			first := make([]byte, len(payload))
			n := DecodePackedLinearFrame(first, payload, flag)
			buff = append(buff, first[:n]...)
			src = payload[n:]
		} else {
			first := make([]byte, len(payload)*2)
			ndst, nsrc := DecodePackedCompactFrame(first, payload, flag&0x7F)
			src = payload[nsrc:]
			buff = append(buff, first[:ndst]...)
		}
	}
	return copy(dst, buff)
}

func DecodePackedLinearFrame(dst, src []byte, flag byte) (n int) {
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

func DecodePackedCompactFrame(dst, src []byte, flag byte) (ndst, nsrc int) {
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

func DecodeFrameMaxLen(x int) int {
	return x * 2
}

func EncodeFrame(dst, src []byte) (n int) {
	for i := 0; i < len(src); n, i = n+1, i+1 {
		if (i % 0x7E) != 0 {
			dst[n] = ^src[i]
		} else {
			remaining := byte(len(src))
			if remaining > 0x7E {
				remaining = 0x7E
			}
			dst[n] = remaining
			n++
			dst[n] = ^src[i]
		}
	}
	return n
}

func EncodeFrameMaxLen(x int) int {
	return x * 2
}
