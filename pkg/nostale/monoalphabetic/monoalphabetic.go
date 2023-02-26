package monoalphabetic

import (
	"bufio"
	"bytes"
	"io"
	"math"

	"github.com/sirupsen/logrus"
)

func NewReader(r io.ByteReader) *Reader {
	return &Reader{
		r:      r,
		mode:   255,
		offset: 4,
	}
}

func NewReaderWithKey(r io.ByteReader, key uint32) *Reader {
	return &Reader{
		r:      r,
		mode:   byte(key & 0xFF),
		offset: byte((key >> 6) & 3),
	}
}

type Reader struct {
	r      io.ByteReader
	mode   byte
	offset byte
}

func (r *Reader) Read(p []byte) (n int, err error) {
	for n = 0; n < len(p); n++ {
		// read the next byte
		c, err := r.r.ReadByte()
		if err != nil {
			return n, err
		}
		// write the decrypted byte to the output
		switch r.mode {
		case 0:
			p[n] = (c - r.offset - 0x40) & 0xFF
		case 1:
			p[n] = (c + r.offset + 0x40) & 0xFF
		case 2:
			p[n] = ((c - r.offset - 0x40) ^ 0xC3) & 0xFF
		case 3:
			p[n] = ((c + r.offset + 0x40) ^ 0xC3) & 0xFF
		default:
			p[n] = (c - 0x0F) & 0xFF
		}
		// if this is the end of a message, return the bytes read so far
		if c == 0xFF {
			return n + 1, nil
		}
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
	logrus.Debugf("monoalphabetic encoded message: %v", binary)
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

func (r *PackedReader) decodePackedChunk(chunk []byte) []byte {
	result := make([]byte, len(chunk))
	for i := 0; i < len(chunk); i += 1 {
		h := int(chunk[i] >> 4)
		l := int(chunk[i] & 0x0F)
		leftByte := r.lookup[h]
		rightByte := r.lookup[l]
		if l != 0 {
			result = append(result, leftByte)
			result = append(result, rightByte)
		} else {
			result = append(result, leftByte)
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
	w io.ByteWriter
}

// NewWriter returns a new Writer reading from r.
//
// To avoid denial of service attacks, the provided bufio.Writer
// should be reading from an io.LimitWriter or similar Writer to bound
// the size of responses.
func NewWriter(w io.ByteWriter) *Writer {
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
	if err := w.w.WriteByte(0xFF); err != nil {
		return nn, err
	}
	return nn + 1, nil
}
