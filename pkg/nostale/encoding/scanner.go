package encoding

import "bytes"

func ScanAuthFrame(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.IndexByte(data, 0xD8); i >= 0 {
		// We have a full frames.
		return i + 1, data[0:i], nil
	}
	// If we're at EOF, we have a final, non-terminated frame. Return it.
	if atEOF {
		return len(data), data, nil
	}
	// Request more data.
	return 0, nil, nil
}

func ScanGatewayFrame(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.IndexByte(data, 0xFF); i >= 0 {
		// We have a full frames.
		return i + 1, data[0:i], nil
	}
	// If we're at EOF, we have a final, non-terminated frame. Return it.
	if atEOF {
		return len(data), data, nil
	}
	// Request more data.
	return 0, nil, nil
}
