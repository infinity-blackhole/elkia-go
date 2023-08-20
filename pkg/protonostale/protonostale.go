package protonostale

import (
	"fmt"
)

var (
	FieldSeparator = []byte(" ")
)

type Unmarshaler interface {
	UnmarshalNosTale([]byte) error
}

type Marshaler interface {
	MarshalNosTale() ([]byte, error)
}

func UnmarshalNosTale(data []byte, v any) error {
	if v, ok := v.(Unmarshaler); ok {
		return v.UnmarshalNosTale(data)
	}
	return fmt.Errorf("invalid payload: %v", v)
}

func MarshalNosTale(v any) ([]byte, error) {
	if v, ok := v.(Marshaler); ok {
		return v.MarshalNosTale()
	}
	return nil, fmt.Errorf("invalid payload: %v", v)
}
