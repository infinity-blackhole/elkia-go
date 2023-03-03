package protonostale

import (
	"bytes"
	"fmt"
	"strconv"

	eventing "github.com/infinity-blackhole/elkia/pkg/api/eventing/v1alpha1"
)

type WorldFrame struct {
	eventing.WorldFrame
}

func (e *WorldFrame) MarshalNosTale() ([]byte, error) {
	var b bytes.Buffer
	if _, err := fmt.Fprintf(&b, "%d ", e.Sequence); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func (e *WorldFrame) UnmarshalNosTale(b []byte) error {
	bs := bytes.SplitN(b, []byte(" "), 2)
	sn, err := strconv.ParseUint(string(bs[0]), 10, 32)
	if err != nil {
		return err
	}
	e.Sequence = uint32(sn)
	return nil
}
