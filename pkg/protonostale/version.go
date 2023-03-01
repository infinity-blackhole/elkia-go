package protonostale

import (
	"bytes"
	"fmt"
	"strconv"
)

func DecodeVersion(b []byte) (string, error) {
	parts := bytes.Split(b, []byte("."))
	if len(parts) != 4 {
		return "", fmt.Errorf("invalid version format")
	}
	major, err := strconv.Atoi(string(parts[0]))
	if err != nil {
		return "", fmt.Errorf("invalid major version: %w", err)
	}
	minor, err := strconv.Atoi(string(parts[1]))
	if err != nil {
		return "", fmt.Errorf("invalid minor version: %w", err)
	}
	patch, err := strconv.Atoi(string(parts[2]))
	if err != nil {
		return "", fmt.Errorf("invalid patch version: %w", err)
	}
	build, err := strconv.Atoi(string(parts[3]))
	if err != nil {
		return "", fmt.Errorf("invalid build version: %w", err)
	}
	return fmt.Sprintf("%d.%d.%d+%d", major, minor, patch, build), nil
}
