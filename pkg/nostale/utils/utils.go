package utils

import (
	"bufio"

	eventing "github.com/infinity-blackhole/elkia/pkg/api/eventing/v1alpha1"
	"github.com/infinity-blackhole/elkia/pkg/protonostale"
	"github.com/sirupsen/logrus"
)

func WriteError(w *bufio.Writer, code eventing.DialogErrorCode, msg string) (n int, err error) {
	if msg != "" {
		logrus.Errorf("error: %s", msg)
		if n, err = protonostale.WriteDialogInfoEvent(w, &eventing.DialogInfoEvent{
			Content: msg,
		}); err != nil {
			return n, err
		}
	}
	if n, err = protonostale.WriteDialogErrorEvent(w, &eventing.DialogErrorEvent{
		Code: code,
	}); err != nil {
		return n, err
	}
	return n, nil
}
