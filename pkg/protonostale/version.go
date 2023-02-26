package protonostale

import (
	"fmt"
	"strconv"
	"strings"
)

func ParseVersion(s string) (string, error) {
	parts := strings.Split(s, ".")
	if len(parts) != 4 {
		return "", fmt.Errorf("invalid version format")
	}
	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return "", fmt.Errorf("invalid major version: %w", err)
	}
	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return "", fmt.Errorf("invalid minor version: %w", err)
	}
	patch, err := strconv.Atoi(parts[2])
	if err != nil {
		return "", fmt.Errorf("invalid patch version: %w", err)
	}
	build, err := strconv.Atoi(parts[3])
	if err != nil {
		return "", fmt.Errorf("invalid build version: %w", err)
	}
	return fmt.Sprintf("%d.%d.%d+%d", major, minor, patch, build), nil
}
