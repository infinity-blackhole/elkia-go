package protonostale

import (
	"fmt"

	eventing "github.com/infinity-blackhole/elkia/pkg/api/eventing/v1alpha1"
)

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
