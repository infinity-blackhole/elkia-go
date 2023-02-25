package monoalphabetic

import (
	"bufio"
	"bytes"
	"math"

	"github.com/sirupsen/logrus"
)

func NewReader(r *bufio.Reader) *Reader {
	return &Reader{
		R:      r,
		mode:   255,
		offset: 4,
	}
}

func NewReaderWithKey(r *bufio.Reader, key uint32) *Reader {
	return &Reader{
		R:      r,
		mode:   byte(key & 0xFF),
		offset: byte((key >> 6) & 3),
	}
}

type Reader struct {
	R      *bufio.Reader
	mode   byte
	offset byte
}

// ReadMessage reads a single message from r,
// eliding the final \n or \r\n from the returned string.
func (r *Reader) ReadMessage() (string, error) {
	msg, err := r.readMessageSlice()
	logrus.Debugf("simple substitution decoded message: %s", msg)
	return string(msg), err
}

// ReadMessageBytes is like ReadMessage but returns a []byte instead of a
// string.
func (r *Reader) ReadMessageBytes() ([]byte, error) {
	msg, err := r.readMessageSlice()
	if msg != nil {
		buf := make([]byte, len(msg))
		copy(buf, msg)
		msg = buf
	}
	return msg, err
}

func (r *Reader) readMessageSlice() ([]byte, error) {
	msg, err := r.R.ReadBytes(0xFF)
	if err != nil {
		return nil, err
	}
	return r.decryptMessage(msg), nil
}

func (r *Reader) decryptMessage(msg []byte) []byte {
	result := make([]byte, len(msg))
	for i, c := range msg {
		result[i] = r.decryptByte(c)
	}
	return result
}

func (r *Reader) decryptByte(c byte) byte {
	switch r.mode {
	case 0:
		return (c - r.offset - 0x40) & 0xFF
	case 1:
		return (c + r.offset + 0x40) & 0xFF
	case 2:
		return ((c - r.offset - 0x40) ^ 0xC3) & 0xFF
	case 3:
		return ((c + r.offset + 0x40) ^ 0xC3) & 0xFF
	default:
		return (c - 0x0F) & 0xFF
	}
}

var lookup = []string{
	"\x00", " ", "-", ".", "0", "1", "2", "3", "4",
	"5", "6", "7", "8", "9", "\n", "\x00",
}

func NewPackedReader(r *Reader) *PackedReader {
	return &PackedReader{
		R:      r,
		lookup: lookup,
	}
}

type PackedReader struct {
	R      *Reader
	lookup []string
}

// ReadMessageSlice reads a single message from r,
// eliding the final \n or \r\n from the returned string.
func (r *PackedReader) ReadMessageSlice() ([]string, error) {
	msgs, err := r.readMessageSlices()
	results := make([]string, len(msgs))
	for _, msg := range msgs {
		results = append(results, string(msg))
	}
	return results, err
}

// ReadMessageSliceBytes is like ReadMessageSlice but returns a [][]byte instead
// of a string.
func (r *PackedReader) ReadMessageSliceBytes() ([][]byte, error) {
	msgs, err := r.readMessageSlices()
	results := make([][]byte, len(msgs))
	for _, msg := range msgs {
		buf := make([]byte, len(msg))
		copy(buf, msg)
		results = append(results, buf)
	}
	return results, err
}

func (r *PackedReader) readMessageSlices() ([][]byte, error) {
	binary, err := r.R.ReadMessageBytes()
	logrus.Debugf("monoalphabetic encoded message: %v", binary)
	if err != nil {
		return nil, err
	}
	return r.unpack(binary), nil
}

func (r *PackedReader) unpack(data []byte) [][]byte {
	packets := make([][]byte, 0)
	parts := bytes.Split(data, []byte{0xFF})
	for _, part := range parts {
		result := r.unpackPart(part)
		packets = append(packets, bytes.Join(result, []byte{}))
	}
	return packets
}

func (r *PackedReader) unpackPart(part []byte) [][]byte {
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

func (r *PackedReader) decodePackedChunk(chunk []byte) []byte {
	result := make([]byte, 0)
	for i := 0; i < len(chunk); i += 1 {
		h := int(chunk[i] >> 4)
		l := int(chunk[i] & 0x0F)
		leftByte := r.lookup[h]
		rightByte := r.lookup[l]
		if l != 0 {
			result = append(result, leftByte...)
			result = append(result, rightByte...)
		} else {
			result = append(result, leftByte...)
		}
	}
	return result
}

func (r *PackedReader) decodeChunk(chunk []byte) []byte {
	result := make([]byte, len(chunk))
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
