package i2p

import (
 "errors"
 "io"
 "strings"
 "testing"
)

type noProgressReader struct{}
func (noProgressReader) Read(_ []byte) (int, error) { return 0, nil }

func TestSAMReadLineNormal(t *testing.T) {
 line, err := readLine(strings.NewReader("HELLO REPLY RESULT=OK\n"), make([]byte, 1))
 if err != nil || line != "HELLO REPLY RESULT=OK\n" { t.Fatalf("line=%q err=%v", line, err) }
}

func TestSAMReadLineRejectsOversizedResponse(t *testing.T) {
 _, err := readLine(strings.NewReader(strings.Repeat("X", maxSAMLineLength+1)+"\n"), make([]byte, 1))
 if !errors.Is(err, errSAMLineTooLong) { t.Fatalf("got %v, want line limit", err) }
}

func TestSAMReadLineRejectsNoProgress(t *testing.T) {
 _, err := readLine(noProgressReader{}, make([]byte, 1))
 if !errors.Is(err, io.ErrNoProgress) { t.Fatalf("got %v, want io.ErrNoProgress", err) }
}

func TestSAMReadLineRejectsEmptyScratch(t *testing.T) {
 _, err := readLine(strings.NewReader("OK\n"), nil)
 if !errors.Is(err, io.ErrShortBuffer) { t.Fatalf("got %v, want io.ErrShortBuffer", err) }
}
