package i2p

import (
 "bufio"
 "net"
 "strings"
 "testing"
 "time"
)

func TestDESTReplyVerification(t *testing.T) {
 for _, tc := range []struct{ line string; valid bool }{
  {"DEST REPLY PUB=public PRIV=private\n", true},
  {"DEST REPLY PUB=public PUB=other PRIV=private\n", false},
  {"DEST REPLY PUB=public PRIV=\n", false},
  {"NAMING REPLY PUB=public PRIV=private\n", false},
 } {
  attrs, err := verifySAMFields(tc.line, "DEST", "REPLY")
  valid := err == nil && attrs["PUB"] != "" && attrs["PRIV"] != ""
  if valid != tc.valid { t.Errorf("%q: valid=%v err=%v", tc.line, valid, err) }
 }
}

func TestDESTGenerateStalledReplyTimesOut(t *testing.T) {
 s := NewKeyfile("")
 a, b := net.Pipe()
 defer a.Close()
 defer b.Close()
 go func() { _, _ = bufio.NewReader(b).ReadString('\n') }()
 done := make(chan error, 1)
 go func() { done <- s.ensure(a) }()
 select {
 case err := <-done:
  if err == nil { t.Fatal("stalled DEST GENERATE succeeded") }
 case <-time.After(samCommandTimeout+2*time.Second):
  t.Fatal("DEST GENERATE did not time out")
 }
}

func TestNAMINGReplyRequiresValue(t *testing.T) {
 for _, line := range []string{
  "NAMING REPLY RESULT=OK\n",
  "NAMING REPLY RESULT=OK VALUE=\n",
  "NAMING REPLY RESULT=KEY_NOT_FOUND NAME=example.i2p\n",
  "NAMING REPLY RESULT=OK RESULT=KEY_NOT_FOUND VALUE=bad\n",
 } {
  attrs, err := verifySAMReply(line, "NAMING", "REPLY")
  if err == nil && strings.TrimSpace(attrs["VALUE"]) != "" {
   t.Errorf("accepted invalid lookup reply: %q", line)
  }
 }
}
