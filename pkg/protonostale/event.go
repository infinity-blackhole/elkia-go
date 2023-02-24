package protonostale

import (
	"bufio"
	"fmt"
	"io"
	"strconv"

	eventing "github.com/infinity-blackhole/elkia/pkg/api/eventing/v1alpha1"
)

type EventReader struct {
	r *bufio.Reader
}

func (r *EventReader) ReadSequence() (uint32, error) {
	return r.ReadUint32()
}

func (r *EventReader) ReadOpCode() (string, error) {
	return r.ReadString()
}

func (r *EventReader) ReadString() (string, error) {
	rf, err := r.ReadField()
	if err != nil {
		return "", err
	}
	return string(rf), nil
}

func (r *EventReader) ReadUint32() (uint32, error) {
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

func (r *EventReader) ReadField() ([]byte, error) {
	b, err := r.r.ReadBytes(' ')
	if err != nil {
		if err == io.EOF {
			return b, nil
		}
		return nil, err
	}
	return b[:len(b)-1], nil
}

func (r *EventReader) ReadPayload() ([]byte, error) {
	b, err := r.r.ReadBytes('\n')
	if err != nil {
		return nil, err
	}
	return b[:len(b)-1], nil
}

func (r *EventReader) Discard(n int) (int, error) {
	return r.r.Discard(n)
}

type EventWriter struct {
	w *bufio.Writer
}

func (w *EventWriter) WriteDialogErrorEvent(msg *eventing.DialogErrorEvent) error {
	_, err := fmt.Fprintf(w.w, "failc %d", msg.Code)
	if err != nil {
		return err
	}
	_, err = fmt.Fprint(w.w, "\n")
	return err
}

func (w *EventWriter) WriteDialogInfoEvent(msg *eventing.DialogInfoEvent) error {
	_, err := fmt.Fprintf(w.w, "info %s", msg.Content)
	if err != nil {
		return err
	}
	_, err = fmt.Fprint(w.w, "\n")
	return err
}
