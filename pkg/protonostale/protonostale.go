package protonostale

var (
	FieldSeparator = []byte(" ")
)

type Unmarshaler interface {
	UnmarshalNosTale([]byte) error
}

type Marshaler interface {
	MarshalNosTale() ([]byte, error)
}
