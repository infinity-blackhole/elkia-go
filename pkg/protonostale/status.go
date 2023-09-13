package protonostale

import (
	"bytes"
	"fmt"
	"strconv"

	eventingpb "go.shikanime.studio/elkia/pkg/api/eventing/v1alpha1"
)

var (
	ErrorOpCode = "failc"
	InfoOpCode  = "info"
)

type ErrorEvent struct {
	*eventingpb.ErrorEvent
}

func (f *ErrorEvent) MarshalNosTale() ([]byte, error) {
	var b bytes.Buffer
	if _, err := fmt.Fprintf(&b, "failc %d", f.Code); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func (f *ErrorEvent) UnmarshalNosTale(b []byte) error {
	f.ErrorEvent = &eventingpb.ErrorEvent{}
	bs := bytes.Split(b, FieldSeparator)
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
	f.Code = eventingpb.Code(code)
	return nil
}

type InfoEvent struct {
	*eventingpb.InfoEvent
}

func (f *InfoEvent) MarshalNosTale() ([]byte, error) {
	var b bytes.Buffer
	if _, err := fmt.Fprintf(&b, "%s %s", InfoOpCode, f.Content); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func (f *InfoEvent) UnmarshalNosTale(b []byte) error {
	f.InfoEvent = &eventingpb.InfoEvent{}
	bs := bytes.Split(b, FieldSeparator)
	if len(bs) != 2 {
		return fmt.Errorf("invalid length: %d", len(bs))
	}
	if string(bs[0]) != InfoOpCode {
		return fmt.Errorf("invalid prefix: %s", string(bs[0]))
	}
	f.Content = string(bs[1])
	return nil
}

func NewStatus(code eventingpb.Code) *Status {
	return &Status{
		&ErrorEvent{
			ErrorEvent: &eventingpb.ErrorEvent{
				Code: code,
			},
		},
	}
}

type Status struct {
	s *ErrorEvent
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
