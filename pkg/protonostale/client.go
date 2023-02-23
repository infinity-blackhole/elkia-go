package protonostale

import (
	"bufio"
	"fmt"

	eventing "github.com/infinity-blackhole/elkia/pkg/api/eventing/v1alpha1"
)

type EventReader struct {
	r *FieldReader
}

func (r *EventReader) ReadOpCode() (string, error) {
	return r.r.ReadString()
}

type Writer struct {
	w *bufio.Writer
}

func (w *Writer) WriteErrorMessageEvent(msg *eventing.ErrorMessageEvent) error {
	_, err := fmt.Fprintf(w.w, "failc %d", msg.Code)
	if err != nil {
		return err
	}
	_, err = fmt.Fprint(w.w, "\n")
	return err
}

func (w *Writer) WriteInfoMessageEvent(msg *eventing.InfoMessageEvent) error {
	_, err := fmt.Fprintf(w.w, "info %s", msg.Content)
	if err != nil {
		return err
	}
	_, err = fmt.Fprint(w.w, "\n")
	return err
}
