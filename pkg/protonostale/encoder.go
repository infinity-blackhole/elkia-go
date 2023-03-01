package protonostale

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"strings"

	eventing "github.com/infinity-blackhole/elkia/pkg/api/eventing/v1alpha1"
)

type DialogErrorEvent struct {
	eventing.DialogErrorEvent
}

func (e *DialogErrorEvent) MarshalNosTale() ([]byte, error) {
	var b bytes.Buffer
	if _, err := fmt.Fprintf(&b, "failc %d", e.Code); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func (e *DialogErrorEvent) UnmarshalNosTale(b []byte) error {
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

type DialogInfoEvent struct {
	eventing.DialogInfoEvent
}

func (e *DialogInfoEvent) MarshalNosTale() ([]byte, error) {
	var b bytes.Buffer
	if _, err := fmt.Fprintf(&b, "%s %s", DialogInfoOpCode, e.Content); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func (e *DialogInfoEvent) UnmarshalNosTale(b []byte) error {
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

type EndpointListEvent struct {
	eventing.EndpointListEvent
}

func (e *EndpointListEvent) MarshalNosTale() ([]byte, error) {
	var buff bytes.Buffer
	if _, err := fmt.Fprintf(&buff, "NsTeST %d ", e.Code); err != nil {
		return nil, err
	}
	for _, m := range e.Endpoints {
		b, err := (&Endpoint{
			Endpoint: eventing.Endpoint{
				Host:      m.Host,
				Port:      m.Port,
				Weight:    m.Weight,
				WorldId:   m.WorldId,
				ChannelId: m.ChannelId,
				WorldName: m.WorldName,
			},
		}).MarshalNosTale()
		if err != nil {
			return nil, err
		}
		if _, err := buff.Write(b); err != nil {
			return nil, err
		}
	}
	if _, err := buff.WriteString("-1:-1:-1:10000.10000.1\n"); err != nil {
		return nil, err
	}
	return buff.Bytes(), nil
}

func (e *EndpointListEvent) UnmarshalNosTale(b []byte) error {
	bs := bytes.Split(b, []byte(" "))
	if len(bs) < 2 {
		return fmt.Errorf("invalid length: %d", len(bs))
	}
	if string(bs[0]) != "NsTeST" {
		return fmt.Errorf("invalid prefix: %s", string(bs[0]))
	}
	code, err := strconv.ParseUint(string(bs[1]), 10, 32)
	if err != nil {
		return err
	}
	e.Code = uint32(code)
	for _, m := range bs[2:] {
		var ep Endpoint
		if err := ep.UnmarshalNosTale(m); err != nil {
			return err
		}
		e.Endpoints = append(e.Endpoints, &ep.Endpoint)
	}
	return nil
}

type Endpoint struct {
	eventing.Endpoint
}

func (e *Endpoint) MarshalNosTale() ([]byte, error) {
	var buff bytes.Buffer
	if _, err := fmt.Fprintf(
		&buff,
		"%s:%s:%d:%d.%d.%s ",
		e.Host,
		e.Port,
		e.Weight,
		e.WorldId,
		e.ChannelId,
		e.WorldName,
	); err != nil {
		return nil, err
	}
	return buff.Bytes(), nil
}

func (e *Endpoint) UnmarshalNosTale(b []byte) error {
	bs := bytes.Split(b, []byte(":"))
	if len(bs) != 5 {
		return fmt.Errorf("invalid length: %d", len(bs))
	}
	e.Host = string(bs[0])
	e.Port = string(bs[1])
	weight, err := strconv.ParseUint(string(bs[2]), 10, 32)
	if err != nil {
		return err
	}
	e.Weight = uint32(weight)
	bs = bytes.Split(bs[3], []byte("."))
	if len(bs) != 3 {
		return fmt.Errorf("invalid length: %d", len(bs))
	}
	worldId, err := strconv.ParseUint(string(bs[0]), 10, 32)
	if err != nil {
		return err
	}
	e.WorldId = uint32(worldId)
	channelId, err := strconv.ParseUint(string(bs[1]), 10, 32)
	if err != nil {
		return err
	}
	e.ChannelId = uint32(channelId)
	e.WorldName = string(bs[2])
	return nil
}

type AuthHandoffLoginEvent struct {
	eventing.AuthHandoffLoginEvent
}

func (e *AuthHandoffLoginEvent) MarshalNosTale() ([]byte, error) {
	var b bytes.Buffer
	bs, err := (&AuthHandoffLoginIdentifierEvent{
		AuthHandoffLoginIdentifierEvent: eventing.AuthHandoffLoginIdentifierEvent{
			Sequence:   e.AuthHandoffLoginEvent.IdentifierEvent.Sequence,
			Identifier: e.AuthHandoffLoginEvent.IdentifierEvent.Identifier,
		},
	}).
		MarshalNosTale()
	if err != nil {
		return nil, err
	}
	if _, err := b.Write(bs); err != nil {
		return nil, err
	}
	bs, err = (&AuthHandoffLoginPasswordEvent{
		AuthHandoffLoginPasswordEvent: eventing.AuthHandoffLoginPasswordEvent{
			Sequence: e.PasswordEvent.Sequence,
			Password: e.PasswordEvent.Password,
		},
	}).
		MarshalNosTale()
	if err != nil {
		return nil, err
	}
	if _, err := b.Write(bs); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func (e *AuthHandoffLoginEvent) UnmarshalNosTale(b []byte) error {
	bs := bytes.Split(b, []byte(" "))
	if len(bs) != 2 {
		return fmt.Errorf("invalid length: %d", len(bs))
	}
	if err := (&AuthHandoffLoginIdentifierEvent{
		AuthHandoffLoginIdentifierEvent: eventing.AuthHandoffLoginIdentifierEvent{
			Sequence:   e.AuthHandoffLoginEvent.IdentifierEvent.Sequence,
			Identifier: e.AuthHandoffLoginEvent.IdentifierEvent.Identifier,
		},
	}).
		UnmarshalNosTale(bs[0]); err != nil {
		return err
	}
	if err := (&AuthHandoffLoginPasswordEvent{
		AuthHandoffLoginPasswordEvent: eventing.AuthHandoffLoginPasswordEvent{
			Sequence: e.AuthHandoffLoginEvent.PasswordEvent.Sequence,
			Password: e.AuthHandoffLoginEvent.PasswordEvent.Password,
		},
	}).
		UnmarshalNosTale(bs[1]); err != nil {
		return err
	}
	return nil
}

type AuthHandoffLoginIdentifierEvent struct {
	eventing.AuthHandoffLoginIdentifierEvent
}

func (e *AuthHandoffLoginIdentifierEvent) MarshalNosTale() ([]byte, error) {
	var b bytes.Buffer
	if _, err := fmt.Fprintf(&b, "%d ", e.Sequence); err != nil {
		return nil, err
	}
	if _, err := b.WriteString(e.Identifier); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func (e *AuthHandoffLoginIdentifierEvent) UnmarshalNosTale(b []byte) error {
	bs := bytes.Split(b, []byte(" "))
	if len(bs) != 2 {
		return fmt.Errorf("invalid length: %d", len(bs))
	}
	sn, err := strconv.ParseUint(string(bs[0]), 10, 32)
	if err != nil {
		return err
	}
	e.Sequence = uint32(sn)
	e.Identifier = string(bs[1])
	return nil
}

type AuthHandoffLoginPasswordEvent struct {
	eventing.AuthHandoffLoginPasswordEvent
}

func (e *AuthHandoffLoginPasswordEvent) MarshalNosTale() ([]byte, error) {
	var b bytes.Buffer
	if _, err := fmt.Fprintf(&b, "%d ", e.Sequence); err != nil {
		return nil, err
	}
	if _, err := b.WriteString(e.Password); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func (e *AuthHandoffLoginPasswordEvent) UnmarshalNosTale(b []byte) error {
	bs := bytes.Split(b, []byte(" "))
	if len(bs) != 2 {
		return fmt.Errorf("invalid length: %d", len(bs))
	}
	sn, err := strconv.ParseUint(string(bs[0]), 10, 32)
	if err != nil {
		return err
	}
	e.Sequence = uint32(sn)
	e.Password = string(bs[1])
	return nil
}

type AuthInteractRequest struct {
	eventing.AuthInteractRequest
}

func (r *AuthInteractRequest) MarshalNosTale() ([]byte, error) {
	var b bytes.Buffer
	switch v := r.Payload.(type) {
	case *eventing.AuthInteractRequest_AuthLoginEvent:
		if _, err := fmt.Fprintf(&b, "%s ", AuthLoginOpCode); err != nil {
			return nil, err
		}
		bs, err := (&AuthLoginEvent{
			AuthLoginEvent: eventing.AuthLoginEvent{
				Identifier:    v.AuthLoginEvent.Identifier,
				Password:      v.AuthLoginEvent.Password,
				ClientVersion: v.AuthLoginEvent.ClientVersion,
			},
		}).MarshalNosTale()
		if err != nil {
			return nil, err
		}
		if _, err := b.Write(bs); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("invalid payload: %v", v)
	}
	return b.Bytes(), nil
}

func (r *AuthInteractRequest) UnmarshalNosTale(b []byte) error {
	bs := bytes.SplitN(b, []byte(" "), 2)
	opcode := string(bs[0])
	switch opcode {
	case AuthLoginOpCode:
		var authLoginEvent AuthLoginEvent
		if err := authLoginEvent.UnmarshalNosTale(bs[1]); err != nil {
			return err
		}
		r.Payload = &eventing.AuthInteractRequest_AuthLoginEvent{
			AuthLoginEvent: &authLoginEvent.AuthLoginEvent,
		}
	default:
		return fmt.Errorf("invalid opcode: %s", opcode)
	}
	return nil
}

type AuthInteractResponse struct {
	eventing.AuthInteractResponse
}

func (r *AuthInteractResponse) MarshalNosTale() ([]byte, error) {
	var b bytes.Buffer
	switch v := r.Payload.(type) {
	case *eventing.AuthInteractResponse_DialogErrorEvent:
		if _, err := fmt.Fprintf(&b, "%s ", DialogErrorOpCode); err != nil {
			return nil, err
		}
		vv := &DialogErrorEvent{
			DialogErrorEvent: eventing.DialogErrorEvent{
				Code: v.DialogErrorEvent.Code,
			},
		}
		bs, err := vv.MarshalNosTale()
		if err != nil {
			return nil, err
		}
		if _, err := b.Write(bs); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("invalid payload: %v", v)
	}
	return b.Bytes(), nil
}

type AuthLoginEvent struct {
	eventing.AuthLoginEvent
}

func (e *AuthLoginEvent) MarshalNosTale() ([]byte, error) {
	var b bytes.Buffer

	if _, err := b.WriteString(e.Identifier); err != nil {
		return nil, err
	}
	password, err := EncodePassword(e.Password)
	if err != nil {
		return nil, err
	}
	if _, err := fmt.Fprintf(&b, " %s", password); err != nil {
		return nil, err
	}
	clientVersion, err := EncodeClientVersion(e.ClientVersion)
	if err != nil {
		return nil, err
	}
	if _, err := fmt.Fprintf(&b, " %s", clientVersion); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func (e *AuthLoginEvent) UnmarshalNosTale(b []byte) error {
	bs := bytes.Split(b, []byte(" "))
	if len(bs) != 4 {
		return fmt.Errorf("invalid length: %d", len(bs))
	}
	e.Identifier = string(bs[1])
	password, err := DecodePassword(bs[2])
	if err != nil {
		return err
	}
	e.Password = password
	clientVersion, err := DecodeClientVersion(bs[3])
	if err != nil {
		return err
	}
	e.ClientVersion = clientVersion
	return nil
}

type AuthHandoffSyncEvent struct {
	eventing.AuthHandoffSyncEvent
}

func (e *AuthHandoffSyncEvent) MarshalNosTale() ([]byte, error) {
	var b bytes.Buffer
	if _, err := fmt.Fprintf(&b, "43%d %d ;;737:584-.37:83898 868 71;481.6; ", e.Sequence, e.Code); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func (e *AuthHandoffSyncEvent) UnmarshalNosTale(b []byte) error {
	bs := bytes.Split(b, []byte(" "))
	if len(bs) != 3 {
		return fmt.Errorf("invalid length: %d", len(bs))
	}
	sn, err := strconv.ParseUint(string(bs[0]), 10, 32)
	if err != nil {
		return err
	}
	e.Sequence = uint32(sn)
	code, err := strconv.ParseUint(string(bs[1]), 10, 32)
	if err != nil {
		return err
	}
	e.Code = uint32(code)
	return nil
}

type ChannelEvent struct {
	eventing.ChannelEvent
}

func (e *ChannelEvent) MarshalNosTale() ([]byte, error) {
	var b bytes.Buffer
	if _, err := fmt.Fprintf(&b, "%d ", e.Sequence); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func (e *ChannelEvent) UnmarshalNosTale(b []byte) error {
	bs := bytes.SplitN(b, []byte(" "), 2)
	sn, err := strconv.ParseUint(string(bs[0]), 10, 32)
	if err != nil {
		return err
	}
	e.Sequence = uint32(sn)
	return nil
}

func DecodePassword(b []byte) (string, error) {
	if len(b)%2 == 0 {
		b = b[3:]
	} else {
		b = b[4:]
	}
	chunks := bytesChunkEvery(b, 2)
	b = make([]byte, 0, len(chunks))
	for i := 0; i < len(chunks); i++ {
		b = append(b, chunks[i][0])
	}
	chunks = bytesChunkEvery(b, 2)
	var result []byte
	for _, chunk := range chunks {
		value, err := strconv.ParseInt(string(chunk), 16, 64)
		if err != nil {
			return "", err
		}
		result = append(result, byte(value))
	}
	return string(result), nil
}

func EncodePassword(s string) (string, error) {
	return "", errors.New("not implemented")
}

func DecodeClientVersion(b []byte) (string, error) {
	fields := bytes.SplitN(b, []byte{'\v'}, 2)
	fields = bytes.Split(fields[1], []byte("."))
	if len(fields) != 4 {
		return "", fmt.Errorf("invalid version: %s", string(b))
	}
	major, err := strconv.ParseUint(string(fields[0]), 10, 32)
	if err != nil {
		return "", err
	}
	minor, err := strconv.ParseUint(string(fields[1]), 10, 32)
	if err != nil {
		return "", err
	}
	patch, err := strconv.ParseUint(string(fields[2]), 10, 32)
	if err != nil {
		return "", err
	}
	build, err := strconv.ParseUint(string(fields[3][:len(fields[3])-1]), 10, 32)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%d.%d.%d+%d", major, minor, patch, build), nil
}

func EncodeClientVersion(version string) (string, error) {
	fields := strings.Split(version, ".")
	if len(fields) != 3 {
		return "", fmt.Errorf("invalid version: %s", version)
	}
	major, err := strconv.ParseUint(fields[0], 10, 32)
	if err != nil {
		return "", err
	}
	minor, err := strconv.ParseUint(fields[1], 10, 32)
	if err != nil {
		return "", err
	}
	fields = strings.Split(fields[2], "+")
	if len(fields) != 2 {
		return "", fmt.Errorf("invalid version: %s", version)
	}
	patch, err := strconv.ParseUint(fields[0], 10, 32)
	if err != nil {
		return "", err
	}
	build, err := strconv.ParseUint(fields[1], 10, 32)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%d.%d.%d.%d", major, minor, patch, build), nil
}
