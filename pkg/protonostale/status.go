package protonostale

import (
	"bytes"
	"fmt"
	"strconv"

	eventing "github.com/infinity-blackhole/elkia/pkg/api/eventing/v1alpha1"
)

var (
	DialogErrorOpCode = "failc"
	DialogInfoOpCode  = "info"
)

type DialogErrorFrame struct {
	eventing.DialogErrorFrame
}

func (e *DialogErrorFrame) MarshalNosTale() ([]byte, error) {
	var b bytes.Buffer
	if _, err := fmt.Fprintf(&b, "failc %d", e.Code); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func (e *DialogErrorFrame) UnmarshalNosTale(b []byte) error {
	bs := bytes.Split(b, []byte(" "))
	if len(bs) != 2 {
		return fmt.Errorf("invalid length: %d", len(bs))
	}
	if string(bs[0]) != "failc" {
		return fmt.Errorf("invalid prefix: %s", string(bs[0]))
	}
	code, err := strconv.ParseUint(string(bs[1]), 10, 32)
	if err != nil {
		return err
	}
	e.Code = eventing.DialogErrorCode(code)
	return nil
}

type DialogInfoFrame struct {
	eventing.DialogInfoFrame
}

func (e *DialogInfoFrame) MarshalNosTale() ([]byte, error) {
	var b bytes.Buffer
	if _, err := fmt.Fprintf(&b, "%s %s", DialogInfoOpCode, e.Content); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func (e *DialogInfoFrame) UnmarshalNosTale(b []byte) error {
	bs := bytes.Split(b, []byte(" "))
	if len(bs) != 2 {
		return fmt.Errorf("invalid length: %d", len(bs))
	}
	if string(bs[0]) != DialogInfoOpCode {
		return fmt.Errorf("invalid prefix: %s", string(bs[0]))
	}
	e.Content = string(bs[1])
	return nil
}

func NewStatus(code eventing.DialogErrorCode) *Status {
	return &Status{
		s: DialogErrorFrame{
			DialogErrorFrame: eventing.DialogErrorFrame{
				Code: code,
			},
		},
	}
}

type Status struct {
	s DialogErrorFrame
}

func (s *Status) Error() string {
	return fmt.Sprintf("status: %v", s.s.Code)
}

func (s *Status) MarshalNosTale() ([]byte, error) {
	return s.s.MarshalNosTale()
}

func (s *Status) UnmarshalNosTale(data []byte) error {
	return s.s.UnmarshalNosTale(data)
}
