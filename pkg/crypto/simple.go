package crypto

type SimpleSubstitution struct {
}

func (s *SimpleSubstitution) Encrypt(plaintext []byte) []byte {
	result := make([]byte, 0, len(plaintext))
	for _, b := range plaintext {
		result = append(result, (b+15)&0xFF)
	}
	return append(result, 0x19)
}

func (s *SimpleSubstitution) Decrypt(ciphertext []byte) []byte {
	result := make([]byte, 0, len(ciphertext))
	for _, b := range ciphertext {
		if b > 14 {
			result = append(result, (b-15)^195)
		} else {
			result = append(result, (255-(14-b))^195)
		}
	}
	return result
}
