package monoalphabetic

import (
	"bufio"
	"bytes"
	"math"

	"github.com/sirupsen/logrus"
)

var charLookup = []string{
	"\x00", " ", "-", ".", "0", "1", "2", "3", "4",
	"5", "6", "7", "8", "9", "\n", "\x00",
}

func NewReader(r *bufio.Reader) *Reader {
	return &Reader{
		R: r,
	}
}

func NewReaderWithCode(r *bufio.Reader, code uint32) *Reader {
	return &Reader{
		R:    r,
		code: &code,
	}
}

type Reader struct {
	R    *bufio.Reader
	code *uint32
}

// ReadMessageSlice reads a single message from r,
// eliding the final \n or \r\n from the returned string.
func (r *Reader) ReadMessageSlice() ([]string, error) {
	msgs, err := r.readMessageSlice()
	results := make([]string, len(msgs))
	for _, msg := range msgs {
		logrus.Debugf("monoalphabetic decoded message: %s", msg)
		results = append(results, string(msg))
	}
	return results, err
}

// ReadMessageSliceBytes is like ReadMessageSlice but returns a [][]byte instead
// of a string.
func (r *Reader) ReadMessageSliceBytes() ([][]byte, error) {
	msgs, err := r.readMessageSlice()
	results := make([][]byte, len(msgs))
	for _, msg := range msgs {
		buf := make([]byte, len(msg))
		copy(buf, msg)
		results = append(results, buf)
	}
	return results, err
}

func (r *Reader) readMessageSlice() ([][]byte, error) {
	binary, err := r.R.ReadBytes(0xFF)
	logrus.Debugf("monoalphabetic encoded message: %s", binary)
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
	if r.code == nil {
		mode := *r.code & 0xFF
		offset := (*r.code >> 6) & 3
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
	return (c - 0x0F) & 0xFF
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
type Writer struct {
	w *bufio.Writer
}

// NewWriter returns a new Writer reading from r.
//
// To avoid denial of service attacks, the provided bufio.Writer
// should be reading from an io.LimitWriter or similar Writer to bound
// the size of responses.
func NewWriter(w *bufio.Writer) *Writer {
	return &Writer{
		w: w,
	}
}

// WriteMessage writes the formatted message.
func (w *Writer) Write(msg []byte) (nn int, err error) {
	for i, b := range msg {
		if i%0x7e != 0 {
			if err := w.w.WriteByte(b); err != nil {
				return nn, err
			}
			nn++
		} else {
			if len(msg)-i > 0x7e {
				if err := w.w.WriteByte(0x7e); err != nil {
					return nn, err
				}
				nn++
			} else {
				if err := w.w.WriteByte(byte(len(msg) - i)); err != nil {
					return nn, err
				}
				nn++
			}
			if err := w.w.WriteByte(b); err != nil {
				return nn, err
			}
			nn++
		}
	}
	if err := w.w.WriteByte(0xff); err != nil {
		return nn, err
	}
	nn++
	return nn, w.w.Flush()
}
