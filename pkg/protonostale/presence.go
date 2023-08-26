package protonostale

import (
	"bytes"
	"fmt"
	"strconv"

	fleet "go.shikanime.studio/elkia/pkg/api/fleet/v1alpha1"
)

type EndpointListEvent struct {
	*fleet.EndpointListEvent
}

func (f *EndpointListEvent) MarshalNosTale() ([]byte, error) {
	var buff bytes.Buffer
	if _, err := fmt.Fprintf(&buff, "NsTeST %d ", f.Code); err != nil {
		return nil, err
	}
	for _, m := range f.Endpoints {
		fields, err := MarshalNosTale(&Endpoint{
			Endpoint: &fleet.Endpoint{
				Host:      m.Host,
				Port:      m.Port,
				Weight:    m.Weight,
				WorldId:   m.WorldId,
				ChannelId: m.ChannelId,
				WorldName: m.WorldName,
			},
		})
		if err != nil {
			return nil, err
		}
		if _, err := buff.Write(fields); err != nil {
			return nil, err
		}
	}
	if _, err := buff.WriteString("-1:-1:-1:10000.10000.1\n"); err != nil {
		return nil, err
	}
	return buff.Bytes(), nil
}

func (f *EndpointListEvent) UnmarshalNosTale(b []byte) error {
	f.EndpointListEvent = &fleet.EndpointListEvent{}
	fields := bytes.Split(b, FieldSeparator)
	if len(fields) < 2 {
		return fmt.Errorf("invalid length: %d", len(fields))
	}
	if string(fields[0]) != "NsTeST" {
		return fmt.Errorf("invalid prefix: %s", string(fields[0]))
	}
	code, err := strconv.ParseUint(string(fields[1]), 10, 32)
	if err != nil {
		return err
	}
	f.Code = uint32(code)
	for _, m := range fields[2:] {
		var ep Endpoint
		if err := ep.UnmarshalNosTale(m); err != nil {
			return err
		}
		f.Endpoints = append(f.Endpoints, ep.Endpoint)
	}
	return nil
}

type Endpoint struct {
	*fleet.Endpoint
}

func (f *Endpoint) MarshalNosTale() ([]byte, error) {
	var buff bytes.Buffer
	if _, err := fmt.Fprintf(
		&buff,
		"%s:%s:%d:%d.%d.%s ",
		f.Host,
		f.Port,
		f.Weight,
		f.WorldId,
		f.ChannelId,
		f.WorldName,
	); err != nil {
		return nil, err
	}
	return buff.Bytes(), nil
}

func (f *Endpoint) UnmarshalNosTale(b []byte) error {
	f.Endpoint = &fleet.Endpoint{}
	fields := bytes.Split(b, []byte(":"))
	if len(fields) != 5 {
		return fmt.Errorf("invalid length: %d", len(fields))
	}
	f.Host = string(fields[0])
	f.Port = string(fields[1])
	weight, err := strconv.ParseUint(string(fields[2]), 10, 32)
	if err != nil {
		return err
	}
	f.Weight = uint32(weight)
	fields = bytes.Split(fields[3], []byte("."))
	if len(fields) != 3 {
		return fmt.Errorf("invalid length: %d", len(fields))
	}
	worldId, err := strconv.ParseUint(string(fields[0]), 10, 32)
	if err != nil {
		return err
	}
	f.WorldId = uint32(worldId)
	channelId, err := strconv.ParseUint(string(fields[1]), 10, 32)
	if err != nil {
		return err
	}
	f.ChannelId = uint32(channelId)
	f.WorldName = string(fields[2])
	return nil
}

type SyncCommand struct {
	*fleet.SyncCommand
}

func (f *SyncCommand) UnmarshalNosTale(b []byte) error {
	f.SyncCommand = &fleet.SyncCommand{}
	fields := bytes.Split(b, FieldSeparator)
	if len(fields) != 3 {
		return fmt.Errorf("invalid length: %d", len(fields))
	}
	sn, err := strconv.ParseUint(string(fields[0][2:]), 10, 32)
	if err != nil {
		return err
	}
	f.Sequence = uint32(sn)
	code, err := strconv.ParseUint(string(fields[1]), 10, 32)
	if err != nil {
		return err
	}
	f.Code = uint32(code)
	return nil
}

type IdentifierCommand struct {
	*fleet.IdentifierCommand
}

func (f *IdentifierCommand) MarshalNosTale() ([]byte, error) {
	var b bytes.Buffer
	if _, err := fmt.Fprintf(&b, "%d ", f.Sequence); err != nil {
		return nil, err
	}
	if _, err := b.WriteString(f.Identifier); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func (f *IdentifierCommand) UnmarshalNosTale(b []byte) error {
	f.IdentifierCommand = &fleet.IdentifierCommand{}
	fields := bytes.Split(b, FieldSeparator)
	if len(fields) != 2 {
		return fmt.Errorf("invalid length: %d", len(fields))
	}
	sn, err := strconv.ParseUint(string(fields[0]), 10, 32)
	if err != nil {
		return err
	}
	f.Sequence = uint32(sn)
	f.Identifier = string(fields[1])
	return nil
}

type PasswordCommand struct {
	*fleet.PasswordCommand
}

func (f *PasswordCommand) MarshalNosTale() ([]byte, error) {
	var b bytes.Buffer
	if _, err := fmt.Fprintf(&b, "%d ", f.Sequence); err != nil {
		return nil, err
	}
	if _, err := b.WriteString(f.Password); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func (f *PasswordCommand) UnmarshalNosTale(b []byte) error {
	f.PasswordCommand = &fleet.PasswordCommand{}
	fields := bytes.Split(b, FieldSeparator)
	if len(fields) != 2 {
		return fmt.Errorf("invalid length: %d", len(fields))
	}
	sn, err := strconv.ParseUint(string(fields[0]), 10, 32)
	if err != nil {
		return err
	}
	f.Sequence = uint32(sn)
	f.Password = string(fields[1])
	return nil
}

type PresenceEvent struct {
	*fleet.PresenceEvent
}

func (f *PresenceEvent) MarshalNosTale() ([]byte, error) {
	return []byte{}, nil
}
