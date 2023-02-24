package protonostale

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
)

func NewVersionReader(r io.Reader) *VersionReader {
	return &VersionReader{
		r: bufio.NewReader(r),
	}
}

type VersionReader struct {
	r *bufio.Reader
}

func (r *VersionReader) ReadVersion() (string, error) {
	if err := r.DiscardPrefix(); err != nil {
		return "", err
	}
	major, err := r.ReadDigit()
	if err != nil {
		return "", err
	}
	minor, err := r.ReadDigit()
	if err != nil {
		return "", err
	}
	patch, err := r.ReadDigit()
	if err != nil {
		return "", err
	}
	build, err := r.ReadDigit()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%d.%d.%d+%d", major, minor, patch, build), nil
}

func (R *VersionReader) DiscardPrefix() error {
	if _, err := R.r.ReadBytes('\v'); err != nil {
		return err
	}
	return nil
}

func (r *VersionReader) ReadDigit() (int, error) {
	var result int
	digit, err := r.r.ReadBytes('.')
	if err != nil {
		if err != io.EOF {
			return 0, err
		}
	}
	result, err = strconv.Atoi(string(digit[:len(digit)-1]))
	if err != nil {
		return 0, err
	}
	return result, nil
}
