package protonostale

import (
	"bufio"
	"fmt"

	eventing "github.com/infinity-blackhole/elkia/pkg/api/eventing/v1alpha1"
)

func NewClientWriter(w *bufio.Writer) *ClientWriter {
	return &ClientWriter{
		w: w,
	}
}

type ClientWriter struct {
	w *bufio.Writer
}

func (w *ClientWriter) WriteFailCodeMessage(msg *eventing.FailureMessage) error {
	_, err := fmt.Fprintf(w.w, "failc %d", msg.Code)
	if err != nil {
		return err
	}
	_, err = fmt.Fprint(w.w, "\n")
	return err
}

func (w *ClientWriter) WriteInfoMessage(msg *eventing.InfoMessage) error {
	_, err := fmt.Fprintf(w.w, "info %s", msg.Content)
	if err != nil {
		return err
	}
	_, err = fmt.Fprint(w.w, "\n")
	return err
}
