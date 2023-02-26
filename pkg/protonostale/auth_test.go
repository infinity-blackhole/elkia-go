package protonostale

import (
	"bufio"
	"bytes"
	"strings"
	"testing"
	"testing/iotest"

	eventing "github.com/infinity-blackhole/elkia/pkg/api/eventing/v1alpha1"
)

func TestParseAuthLoginEvent(t *testing.T) {
	input := "2503350 admin 9827F3538326B33722633327E4 006666A8\v0.9.3.3086"
	expected := &eventing.AuthLoginEvent{
		Identifier:    "admin",
		Password:      "s3cr3t",
		ClientVersion: "0.9.3+3086",
	}
	r := bufio.NewReader(
		iotest.NewReadLogger(t.Name(), strings.NewReader(input)),
	)
	result, err := ReadAuthLoginEvent(r)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if result.Identifier != expected.Identifier {
		t.Errorf("Expected %v, got %v", expected.Identifier, result.Identifier)
	}
	if result.Password != expected.Password {
		t.Errorf("Expected %v, got %v", expected.Password, result.Password)
	}
	if result.ClientVersion != expected.ClientVersion {
		t.Errorf("Expected %v, got %v", expected.ClientVersion, result.ClientVersion)
	}
}

func TestReadPassword(t *testing.T) {
	input := "2EB6A196E4B60D96A9267E"
	expected := "admin"
	r := bufio.NewReader(
		iotest.NewReadLogger(t.Name(), strings.NewReader(input)),
	)
	result, err := ReadPassword(r)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if result != expected {
		t.Errorf("Expected %v, got %v", expected, result)
	}

	input = "1BE97B527A306B597A2"
	expected = "user"
	r = bufio.NewReader(
		iotest.NewReadLogger(t.Name(), strings.NewReader(input)),
	)
	result, err = ReadPassword(r)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if result != expected {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestReadVersion(t *testing.T) {
	input := []byte{
		48, 48, 57, 55, 54, 55, 69, 67, 11, 48, 46, 57, 46, 51, 46, 51, 48,
		56, 54,
	}
	expected := "0.9.3+3086"
	r := bufio.NewReader(
		iotest.NewReadLogger(t.Name(), bytes.NewBuffer(input)),
	)
	result, err := ReadVersion(r)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if result != expected {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestWriteGatewayListEvent(t *testing.T) {
	input := &eventing.GatewayListEvent{
		Key: 1,
		Gateways: []*eventing.Gateway{
			{
				Host:      "127.0.0.1",
				Port:      "4124",
				Weight:    0,
				WorldId:   1,
				ChannelId: 1,
				WorldName: "Test",
			},
			{
				Host:      "127.0.0.1",
				Port:      "4125",
				Weight:    0,
				WorldId:   1,
				ChannelId: 2,
				WorldName: "Test",
			},
		},
	}
	expected := "NsTeST 1 127.0.0.1:4124:0:1.1.Test 127.0.0.1:4125:0:1.2.Test -1:-1:-1:10000.10000.1"
	var result bytes.Buffer
	w := iotest.NewWriteLogger(t.Name(), &result)
	n, err := WriteGatewayListEvent(w, input)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if n != len(expected) {
		t.Errorf("Expected %v, got %v", len(expected), n)
	}
	if result.String() != expected {
		t.Errorf("Expected %v, got %v", expected, result.String())
	}
}
