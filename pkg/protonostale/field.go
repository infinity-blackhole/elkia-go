package protonostale

import (
	"bufio"
	"io"
	"strconv"
)

func NewFieldReader(r io.Reader) *FieldReader {
	return &FieldReader{
		r: bufio.NewReader(r),
	}
}

type FieldReader struct {
	r *bufio.Reader
}

func (r *FieldReader) ReadString() (string, error) {
	rf, err := r.ReadField()
	if err != nil {
		return "", err
	}
	return string(rf), nil
}

func (r *FieldReader) ReadUint32() (uint32, error) {
	rf, err := r.ReadField()
	if err != nil {
		return 0, err
	}
	key, err := strconv.ParseUint(string(rf), 10, 32)
	if err != nil {
		return 0, err
	}
	return uint32(key), nil
}

func (r *FieldReader) ReadField() ([]byte, error) {
	b, err := r.r.ReadBytes(' ')
	if err != nil {
		return nil, err
	}
	return b[:len(b)-1], nil
}

func (r *FieldReader) ReadPayload() ([]byte, error) {
	b, err := r.r.ReadBytes('\n')
	if err != nil {
		return nil, err
	}
	return b[:len(b)-1], nil
}

func (r *FieldReader) Discard(n int) (int, error) {
	return r.r.Discard(n)
}
