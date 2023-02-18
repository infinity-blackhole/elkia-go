package crypto

import (
	"bufio"
	"bytes"
	"math"
)

var charLookup = []string{
	"\x00", " ", "-", ".", "0", "1", "2", "3", "4",
	"5", "6", "7", "8", "9", "\n", "\x00",
}

type ServerReader struct {
	R   *bufio.Reader
	key uint32
}

func NewServerReader(r *bufio.Reader, key uint32) *ServerReader {
	return &ServerReader{
		R:   r,
		key: key,
	}
}

func (r *ServerReader) ReadMessage() ([][]byte, error) {
	if r.key == 0 {
		return r.ReadSessionMessage()
	}
	return r.ReadChannelMessage()
}

func (r *ServerReader) ReadSessionMessage() ([][]byte, error) {
	binary, err := r.R.ReadBytes(0xFF)
	if err != nil {
		return nil, err
	}
	return r.unpack(r.decryptMessage(binary)), nil
}

func (r *ServerReader) ReadChannelMessage() ([][]byte, error) {
	binary, err := r.R.ReadBytes(0xFF)
	if err != nil {
		return nil, err
	}
	return r.unpack(r.decryptMessage(binary)), nil
}

func (r *ServerReader) decryptMessage(msg []byte) []byte {
	result := make([]byte, 0)
	for _, c := range msg {
		result = append(result, r.decryptByte(c))
	}
	return result
}

func (r *ServerReader) decryptByte(c byte) byte {
	mode := r.key & 0xFF
	offset := (r.key >> 6) & 3
	switch mode {
	case 0:
		return (c - byte(offset) - 0x40) & 0xFF
	case 1:
		return (c + byte(offset) + 0x40) & 0xFF
	case 2:
		return ((c - byte(offset) - 0x40) ^ 0xC3) & 0xFF
	case 3:
		return ((c + byte(offset) + 0x40) ^ 0xC3) & 0xFF
	default:
		return (c - 0x0F) & 0xFF
	}
}

func (r *ServerReader) unpack(data []byte) [][]byte {
	packets := make([][]byte, 0)
	parts := bytes.Split(data, []byte{0xFF})
	for _, part := range parts {
		packet := r.doUnpack(part, make([][]byte, 0))
		packets = append(packets, packet)
	}
	return packets
}

func (r *ServerReader) doUnpack(data []byte, result [][]byte) []byte {
	if len(data) == 0 {
		r.reverseBytes(result)
		return bytes.Join(result, []byte{})
	}

	byteVal := data[0]
	rest := data[1:]

	isPacked := (byteVal & 0x80) > 0
	tmpLen := byteVal & 0x7F
	var length int
	if isPacked {
		length = int(math.Ceil(float64(tmpLen) / 2))
	} else {
		length = int(tmpLen)
	}

	chunk := rest[:length]
	next := rest[length:]

	decodedChunk := r.decodeChunk(chunk, isPacked)
	result = append(result, decodedChunk)

	return r.doUnpack(next, result)
}

func (r *ServerReader) decodeChunk(chunk []byte, isPacked bool) []byte {
	decodedChunk := make([]byte, 0)
	if !isPacked {
		for _, c := range chunk {
			decodedChunk = append(decodedChunk, c^0xFF)
		}
	} else {
		for i := 0; i < len(chunk); i += 2 {
			h := int(chunk[i] >> 4)
			l := int(chunk[i] & 0x0F)
			leftByte := charLookup[h]
			rightByte := charLookup[l]
			if l != 0 {
				decodedChunk = append(decodedChunk, leftByte...)
				decodedChunk = append(decodedChunk, rightByte...)
			} else {
				decodedChunk = append(decodedChunk, leftByte...)
			}
		}
	}
	return decodedChunk
}

func (r *ServerReader) reverseBytes(data [][]byte) {
	for i, j := 0, len(data)-1; i < j; i, j = i+1, j-1 {
		data[i], data[j] = data[j], data[i]
	}
}

// A Writer implements convenience methods for reading messages
// from a NosTale protocol network connection.
type MonoAlphabeticWriter struct {
	W *bufio.Writer
}

// NewWriter returns a new Writer reading from r.
//
// To avoid denial of service attacks, the provided bufio.Writer
// should be reading from an io.LimitWriter or similar Writer to bound
// the size of responses.
func NewMonoAlphabeticWriter(r *bufio.Writer) *MonoAlphabeticWriter {
	return &MonoAlphabeticWriter{}
}

// WriteMessage writes the formatted message.
func (w *MonoAlphabeticWriter) WriteMessage(msg []byte) error {
	var result []byte
	for i, b := range msg {
		if i%0x7e != 0 {
			result = append(result, b)
		} else {
			var rest int
			if len(msg)-i > 0x7e {
				rest = 0x7e
			} else {
				rest = len(msg) - i
			}
			result = append(result, []byte{byte(rest), b}...)
		}
	}
	if _, err := w.W.Write(append(result, 0xff)); err != nil {
		return err
	}
	return w.W.Flush()
}
