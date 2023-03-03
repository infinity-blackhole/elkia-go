package protonostale

import (
	"bytes"
	"fmt"
	"strconv"

	eventing "github.com/infinity-blackhole/elkia/pkg/api/eventing/v1alpha1"
)

var (
	ErrorOpCode = "failc"
	InfoOpCode  = "info"
)

type ErrorFrame struct {
	eventing.ErrorFrame
}

func (e *ErrorFrame) MarshalNosTale() ([]byte, error) {
	var b bytes.Buffer
	if _, err := fmt.Fprintf(&b, "failc %d", e.Code); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func (e *ErrorFrame) UnmarshalNosTale(b []byte) error {
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
	e.Code = eventing.Code(code)
	return nil
}

type InfoFrame struct {
	eventing.InfoFrame
}

func (e *InfoFrame) MarshalNosTale() ([]byte, error) {
	var b bytes.Buffer
	if _, err := fmt.Fprintf(&b, "%s %s", InfoOpCode, e.Content); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func (e *InfoFrame) UnmarshalNosTale(b []byte) error {
	bs := bytes.Split(b, []byte(" "))
	if len(bs) != 2 {
		return fmt.Errorf("invalid length: %d", len(bs))
	}
	if string(bs[0]) != InfoOpCode {
		return fmt.Errorf("invalid prefix: %s", string(bs[0]))
	}
	e.Content = string(bs[1])
	return nil
}

func NewStatus(code eventing.Code) *Status {
	return &Status{
		s: ErrorFrame{
			ErrorFrame: eventing.ErrorFrame{
				Code: code,
			},
		},
	}
}

type Status struct {
	s ErrorFrame
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
