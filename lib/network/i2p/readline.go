package i2p

import (
	"errors"
	"io"
)

const maxSAMLineLength = 64 * 1024

var errSAMLineTooLong = errors.New("SAM response line exceeds limit")

// readLine reads one protocol line without allowing an unbounded allocation.
// SAM callers supply a one-byte scratch buffer.
func readLine(r io.Reader, buff []byte) (line string, err error) {
	if len(buff) == 0 {
		return "", io.ErrShortBuffer
	}
	data := make([]byte, 0, 256)
	for len(data) < maxSAMLineLength {
		n, readErr := r.Read(buff[:1])
		if n > 0 {
			data = append(data, buff[0])
			if buff[0] == '\n' {
				return string(data), nil
			}
		}
		if readErr != nil {
			return string(data), readErr
		}
		if n == 0 {
			return string(data), io.ErrNoProgress
		}
	}
	return string(data), errSAMLineTooLong
}
