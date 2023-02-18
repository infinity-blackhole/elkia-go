package protonostale

import (
	"bufio"
	"strconv"
)

func NewFieldReader(r *bufio.Reader) *FieldReader {
	return &FieldReader{
		r: r,
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
	return r.r.ReadBytes(' ')
}

func (r *FieldReader) ReadPayload() ([]byte, error) {
	return r.r.ReadBytes('\n')
}
