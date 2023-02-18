package monoalphabetic

import (
	"bufio"
	"bytes"
	"math"
)

var charLookup = []string{
	"\x00", " ", "-", ".", "0", "1", "2", "3", "4",
	"5", "6", "7", "8", "9", "\n", "\x00",
}

type Reader struct {
	R   *bufio.Reader
	key uint32
}

func NewReader(r *bufio.Reader) *Reader {
	return &Reader{
		R: r,
	}
}

// ReadMessage reads a single message from r,
// eliding the final \n or \r\n from the returned string.
func (r *Reader) ReadMessage() ([]string, error) {
	msgs, err := r.readMessageSlice()
	results := make([]string, len(msgs))
	for _, msg := range msgs {
		results = append(results, string(msg))
	}
	return results, err
}

// ReadMessageBytes is like ReadMessage but returns a []byte instead of a
// string.
func (r *Reader) ReadMessageBytes() ([][]byte, error) {
	msgs, err := r.readMessageSlice()
	results := make([][]byte, len(msgs))
	for _, msg := range msgs {
		if msg != nil {
			buf := make([]byte, len(msg))
			copy(buf, msg)
			results = append(results, buf)
		}
	}
	return results, err
}

func (r *Reader) readMessageSlice() ([][]byte, error) {
	if r.key == 0 {
		return r.readSessionMessage()
	}
	return r.readChannelMessage()
}

func (r *Reader) readSessionMessage() ([][]byte, error) {
	binary, err := r.R.ReadBytes(0xFF)
	if err != nil {
		return nil, err
	}
	return r.unpack(r.decryptMessage(binary)), nil
}

func (r *Reader) readChannelMessage() ([][]byte, error) {
	binary, err := r.R.ReadBytes(0xFF)
	if err != nil {
		return nil, err
	}
	return r.unpack(r.decryptMessage(binary)), nil
}

func (r *Reader) decryptMessage(msg []byte) []byte {
	result := make([]byte, 0)
	for _, c := range msg {
		result = append(result, r.decryptByte(c))
	}
	return result
}

func (r *Reader) decryptByte(c byte) byte {
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

func (r *Reader) unpack(data []byte) [][]byte {
	packets := make([][]byte, 0)
	parts := bytes.Split(data, []byte{0xFF})
	for _, part := range parts {
		result := r.unpackPart(part)
		packets = append(packets, bytes.Join(result, []byte{}))
	}
	return packets
}

func (r *Reader) unpackPart(part []byte) [][]byte {
	result := make([][]byte, 0)
	for len(part) != 0 {
		byteVal := part[0]
		rest := part[1:]
		isPacked := (byteVal & 0x80) > 0
		tmpLen := byteVal & 0x7F
		var length int
		if isPacked {
			length = int(math.Ceil(float64(tmpLen) / 2))
		} else {
			length = int(tmpLen)
		}
		chunk := rest[:length]
		part = rest[length:]
		var decodedChunk []byte
		if isPacked {
			decodedChunk = r.decodePackedChunk(chunk)
		} else {
			decodedChunk = r.decodeChunk(chunk)
		}
		result = append(result, decodedChunk)
	}
	return result
}

func (r *Reader) decodePackedChunk(chunk []byte) []byte {
	result := make([]byte, 0)
	for i := 0; i < len(chunk); i += 2 {
		h := int(chunk[i] >> 4)
		l := int(chunk[i] & 0x0F)
		leftByte := charLookup[h]
		rightByte := charLookup[l]
		if l != 0 {
			result = append(result, leftByte...)
			result = append(result, rightByte...)
		} else {
			result = append(result, leftByte...)
		}
	}
	return result
}

func (r *Reader) decodeChunk(chunk []byte) []byte {
	result := make([]byte, 0)
	for _, c := range chunk {
		result = append(result, c^0xFF)
	}
	return result
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
