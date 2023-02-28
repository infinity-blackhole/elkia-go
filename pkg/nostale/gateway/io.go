package gateway

import (
	"bufio"
	"bytes"
	"errors"
	"math"

	"github.com/sirupsen/logrus"
)

func NewHandoffReader(r *bufio.Reader) *HandoffReader {
	return &HandoffReader{
		r: r,
	}
}

type HandoffReader struct {
	r *bufio.Reader
}

func (r *HandoffReader) Read(p []byte) (n int, err error) {
	for n = 0; len(p) > n*2; n++ {
		c, err := r.r.ReadByte()
		if err != nil {
			return n, err
		}
		first, second, err := r.decodeBytePair(c)
		if err != nil {
			return n, err
		}
		p[n*2] = first
		p[n*2+1] = second
	}
	return n, nil
}

func (r *HandoffReader) decodeBytePair(b byte) (byte, byte, error) {
	first_byte := b - 0xF
	second_byte := first_byte & 0xF0
	second_key := second_byte >> 0x4
	first_key := first_byte - second_byte
	return r.decodeByte(second_key), r.decodeByte(first_key), nil
}

func (*HandoffReader) decodeByte(b byte) byte {
	switch b {
	case 0, 1:
		return ' '
	case 2:
		return '-'
	case 3:
		return '.'
	default:
		return 0x2C + b
	}
}

func NewReader(r *bufio.Reader, key uint32) *Reader {
	return &Reader{
		r:      r,
		mode:   byte(key >> 6 & 0x03),
		offset: byte(key&0xFF + 0x40&0xFF),
	}
}

type Reader struct {
	r      *bufio.Reader
	mode   byte
	offset byte
}

func (r *Reader) ReadByte() (byte, error) {
	c, err := r.r.ReadByte()
	if err != nil {
		return c, err
	}
	switch r.mode {
	case 0:
		return c - r.offset, nil
	case 1:
		return c + r.offset, nil
	case 2:
		return (c - r.offset) ^ 0xC3, nil
	case 3:
		return (c + r.offset) ^ 0xC3, nil
	default:
		return c, errors.New("invalid mode")
	}
}

func (r *Reader) Read(p []byte) (n int, err error) {
	for n = 0; n < len(p); n++ {
		c, err := r.ReadByte()
		if err != nil {
			return n, err
		}
		p[n] = c
	}
	return n, nil
}

var lookup = []byte{
	'\x00', ' ', '-', '.', '0', '1', '2', '3', '4',
	'5', '6', '7', '8', '9', '\n', '\x00',
}

func NewPackedReader(r *bufio.Reader) *PackedReader {
	return &PackedReader{
		r:      r,
		lookup: lookup,
	}
}

type PackedReader struct {
	r      *bufio.Reader
	lookup []byte
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
	binary, err := r.r.ReadBytes(0xFF)
	if err != nil {
		return nil, err
	}
	logrus.Debugf("gatewayio encoded message: %v", binary)
	return r.unpack(binary), nil
}

func (r *PackedReader) unpack(data []byte) [][]byte {
	parts := bytes.Split(data, []byte{0xFF})
	packets := make([][]byte, len(parts))
	for i, part := range parts {
		packets[i] = bytes.Join(r.unpackPart(part), []byte{})
	}
	return packets
}

func (r *PackedReader) unpackPart(binary []byte) [][]byte {
	var result [][]byte
	for len(binary) > 0 {
		head := binary[0]
		binary = binary[1:]
		isPacked := head&0x80 > 0
		tmpLen := head & 0x7F
		var partLen int
		if isPacked {
			partLen = int(math.Ceil(float64(tmpLen) / 2))
		} else {
			partLen = int(tmpLen)
		}

		if partLen == 0 {
			continue
		}

		var chunk []byte
		if partLen <= len(r.lookup) {
			chunk = append(chunk, r.lookup[partLen-1])
		} else {
			chunk = make([]byte, partLen)
			for i := 0; i < partLen; i++ {
				chunk[i] = byte(0xFF)
			}
		}

		if isPacked {
			var decodedChunk []byte
			for i := 0; i < partLen/2; i++ {
				b := binary[0]
				binary = binary[1:]
				hi := (b >> 4) & 0xF
				lo := b & 0xF
				decodedChunk = append(decodedChunk, chunk[hi], chunk[lo])
			}
			if partLen%2 == 1 {
				decodedChunk = append(decodedChunk, chunk[binary[0]>>4])
			}
			chunk = decodedChunk
		}

		result = append(result, chunk)
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

// Write writes the formatted message.
func (w *Writer) Write(msg []byte) (n int, err error) {
	for n = 0; n < len(msg); n++ {
		if (n % 0x7E) != 0 {
			if err := w.w.WriteByte(^msg[n]); err != nil {
				return n, err
			}
		} else {
			remaining := byte(len(msg) - n)
			if remaining > 0x7E {
				remaining = 0x7E
			}
			if err := w.w.WriteByte(remaining); err != nil {
				return n, err
			}
			if err := w.w.WriteByte(^msg[n]); err != nil {
				return n, err
			}
		}
	}
	return n, w.w.Flush()
}
