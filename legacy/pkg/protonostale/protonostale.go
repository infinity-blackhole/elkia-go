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
	switch v := v.(type) {
	case Unmarshaler:
		return v.UnmarshalNosTale(data)
	case *string:
		*v = string(data)
		return nil
	case *[]byte:
		*v = data
		return nil
	default:
		return fmt.Errorf("invalid payload: %v", v)
	}
}

func MarshalNosTale(v any) ([]byte, error) {
	switch v := v.(type) {
	case []byte:
		return v, nil
	case string:
		return []byte(v), nil
	case Marshaler:
		return v.MarshalNosTale()
	default:
		return nil, fmt.Errorf("invalid payload: %v", v)
	}
}
