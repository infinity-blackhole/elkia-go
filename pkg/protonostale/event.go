package protonostale

import (
	"bufio"
	"fmt"

	eventing "github.com/infinity-blackhole/elkia/pkg/api/eventing/v1alpha1"
)

type EventReader struct {
	FieldReader *FieldReader
}

func (r *EventReader) ReadSequence() (uint32, error) {
	return r.FieldReader.ReadUint32()
}

func (r *EventReader) ReadOpCode() (string, error) {
	return r.FieldReader.ReadString()
}

type Writer struct {
	w *bufio.Writer
}

func (w *Writer) WriteDialogErrorEvent(msg *eventing.DialogErrorEvent) error {
	_, err := fmt.Fprintf(w.w, "failc %d", msg.Code)
	if err != nil {
		return err
	}
	_, err = fmt.Fprint(w.w, "\n")
	return err
}

func (w *Writer) WriteDialogInfoEvent(msg *eventing.DialogInfoEvent) error {
	_, err := fmt.Fprintf(w.w, "info %s", msg.Content)
	if err != nil {
		return err
	}
	_, err = fmt.Fprint(w.w, "\n")
	return err
}
