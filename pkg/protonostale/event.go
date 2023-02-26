package protonostale

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strconv"

	eventing "github.com/infinity-blackhole/elkia/pkg/api/eventing/v1alpha1"
)

func ReadSequence(r *bufio.Reader) (uint32, error) {
	return ReadUint32(r)
}

func ReadOpCode(r *bufio.Reader) (string, error) {
	return ReadString(r)
}

func ReadString(r *bufio.Reader) (string, error) {
	rf, err := ReadField(r)
	if err != nil {
		return "", err
	}
	return string(rf), nil
}

func ReadUint32(r *bufio.Reader) (uint32, error) {
	rf, err := ReadField(r)
	if err != nil {
		return 0, err
	}
	key, err := strconv.ParseUint(string(rf), 10, 32)
	if err != nil {
		return 0, err
	}
	return uint32(key), nil
}

func ReadField(r *bufio.Reader) ([]byte, error) {
	b, err := r.ReadBytes(' ')
	if err != nil {
		if err == io.EOF {
			return b, nil
		}
		return nil, err
	}
	return b[:len(b)-1], nil
}

func ReadPayload(r *bufio.Reader) ([]byte, error) {
	b, err := r.ReadBytes('\n')
	if err != nil {
		return nil, err
	}
	return b[:len(b)-1], nil
}

func WriteDialogErrorEvent(
	w io.Writer,
	msg *eventing.DialogErrorEvent,
) (n int, err error) {
	var b bytes.Buffer
	if _, err := fmt.Fprintf(&b, "failc %d", msg.Code); err != nil {
		return n, err
	}
	if err = b.WriteByte('\n'); err != nil {
		return n, err
	}
	return w.Write(b.Bytes())
}

func WriteDialogInfoEvent(
	w io.Writer,
	msg *eventing.DialogInfoEvent,
) (n int, err error) {
	var b bytes.Buffer
	if _, err := fmt.Fprintf(&b, "info %s", msg.Content); err != nil {
		return n, err
	}
	if err = b.WriteByte('\n'); err != nil {
		return n, err
	}
	return w.Write(b.Bytes())
}
