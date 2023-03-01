package protonostale

import (
	"bufio"
	"bytes"
	"fmt"
	"strconv"

	eventing "github.com/infinity-blackhole/elkia/pkg/api/eventing/v1alpha1"
)

func DecodeUint(b []byte) (uint32, error) {
	code, err := strconv.ParseUint(string(b), 10, 32)
	if err != nil {
		return 0, err
	}
	return uint32(code), nil
}

func WriteEvent(w *bufio.Writer, b []byte) (n int, err error) {
	nn, err := w.Write(b)
	if err != nil {
		return n, err
	}
	n += nn
	if err = w.WriteByte('\n'); err != nil {
		return n, err
	}
	return n + 1, nil
}

func WriteDialogErrorEvent(
	w *bufio.Writer,
	msg *eventing.DialogErrorEvent,
) (n int, err error) {
	var b bytes.Buffer
	if _, err := fmt.Fprintf(&b, "failc %d", msg.Code); err != nil {
		return n, err
	}
	return WriteEvent(w, b.Bytes())
}

func WriteDialogInfoEvent(
	w *bufio.Writer,
	msg *eventing.DialogInfoEvent,
) (n int, err error) {
	var b bytes.Buffer
	if _, err := fmt.Fprintf(&b, "info %s", msg.Content); err != nil {
		return n, err
	}
	return WriteEvent(w, b.Bytes())
}
