package protonostale

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
)

func ReadVersion(r *bufio.Reader) (string, error) {
	if _, err := ReadVersionPrefix(r); err != nil {
		return "", err
	}
	major, err := ReadVersionDigit(r)
	if err != nil {
		return "", err
	}
	minor, err := ReadVersionDigit(r)
	if err != nil {
		return "", err
	}
	patch, err := ReadVersionDigit(r)
	if err != nil {
		return "", err
	}
	build, err := ReadVersionDigit(r)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%d.%d.%d+%d", major, minor, patch, build), nil
}

func ReadVersionPrefix(r *bufio.Reader) ([]byte, error) {
	return r.ReadBytes('\v')
}

func ReadVersionDigit(r *bufio.Reader) (int, error) {
	digit, err := r.ReadBytes('.')
	offset := len(digit) - 1
	if err != nil {
		if err != io.EOF {
			return 0, err
		} else {
			offset = len(digit)
		}
	}
	return strconv.Atoi(string(digit[:offset]))
}
