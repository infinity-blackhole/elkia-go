package utils

import (
	"bufio"
	"fmt"

	eventing "github.com/infinity-blackhole/elkia/pkg/api/eventing/v1alpha1"
	"github.com/infinity-blackhole/elkia/pkg/protonostale"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TranslateError(err error) eventing.DialogErrorCode {
	if err, ok := status.FromError(err); ok {
		switch err.Code() {
		case codes.InvalidArgument:
			return eventing.DialogErrorCode_BAD_CASE
		}
	}
	return eventing.DialogErrorCode_UNEXPECTED_ERROR
}

func WriteError(w *bufio.Writer, code eventing.DialogErrorCode, msg string) (n int, err error) {
	if msg != "" {
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

func WriteErrorf(w *bufio.Writer, code eventing.DialogErrorCode, msg string, a ...any) (n int, err error) {
	return WriteError(w, code, fmt.Sprintf(msg, a...))
}
