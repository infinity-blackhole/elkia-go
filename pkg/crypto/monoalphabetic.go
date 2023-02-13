package crypto

import (
	"bytes"
	"math"
)

var (
	permutationMatrix = []string{
		"\x00", " ", "-", ".", "0", "1", "2", "3", "4",
		"5", "6", "7", "8", "9", "\n", "\x00",
	}
)

type MonoAlphabetic struct {
	key *uint32
}

func NewMonoAlphabetic(key uint32) *MonoAlphabetic {
	return &MonoAlphabetic{
		key: key,
	}
}

func (c MonoAlphabetic) Encrypt(plaintext []byte) []byte {
	var result []byte
	for i, b := range plaintext {
		if i%0x7e != 0 {
			result = append(result, b)
		} else {
			var rest int
			if len(plaintext)-i > 0x7e {
				rest = 0x7e
			} else {
				rest = len(plaintext) - i
			}
			result = append(result, []byte{byte(rest), b}...)
		}
	}
	return append(result, 0xff)
}

func (c MonoAlphabetic) Decrypt(plaintext []byte) []byte {
	return []byte(c.unpack(c.decryptBytes(plaintext, -1, -1)))
}

func (c MonoAlphabetic) DecryptWithKey(plaintext []byte, sessionKey uint32) []byte {
	decryptionType := (sessionKey >> 6) & 3
	offset := sessionKey & 0xff
	return []byte(c.unpack(c.decryptBytes(plaintext, offset, decryptionType)))
}

func (c MonoAlphabetic) decryptBytes(plaintext []byte, offset byte, decryptionType int) []byte {
	var result []byte
	for _, c := range plaintext {
		switch decryptionType {
		case 0:
			result = append(result, (c-offsetByte-0x40)&0xff)
		case 1:
			result = append(result, (c+offsetByte+0x40)&0xff)
		case 2:
			result = append(result, (c-offsetByte-0x40)^0xc3&0xff)
		case 3:
			result = append(result, (c+offsetByte+0x40)^0xc3&0xff)
		default:
			result = append(result, (c-0x0f)&0xff)
		}
	}
	return result
}

func (c MonoAlphabetic) unpack(binaryData []byte) []byte {
	var result []byte
	for _, data := range bytes.Split(binaryData, []byte{0xff}) {
		result = append(result, c.unpackBytes(data, result)...)
	}
	return result
}

func (c MonoAlphabetic) unpackBytes(data []byte, result []byte) []byte {
	if data == "" {
		result = reverse(result)
		return result
	}
	byte := data[0]
	rest := data[1:]

	packed := byte&0x80 > 0
	len := byte & 0x7F

	if packed {
		len = int(math.Ceil(float64(len) / 2))
	}

	pack := rest[:len]
	rest = rest[len:]

	if packed {
		temp := ""
		for i := 0; i < len; i += 4 {
			h := pack[i]
			l := pack[i+1]

			leftByte := permutationMatrix[h]
			rightByte := permutationMatrix[l]

			if l != 0 {
				temp += string(leftByte) + string(rightByte)
			} else {
				temp += string(leftByte)
			}
		}
		pack = temp
	} else {
		temp := ""
		for i := 0; i < len; i++ {
			c := pack[i]
			temp += string(c ^ 0xFF)
		}
		pack = temp
	}

	return c.unpackBytes(rest, append(result, pack))
}

func reverse(result []byte) []byte {
	for i := 0; i < len(result)/2; i++ {
		result[i], result[len(result)-i-1] = result[len(result)-i-1], result[i]
	}
	return result
}
