package i2p

import "testing"

func TestVerifySAMReply(t *testing.T) {
 tests := []struct{
  name, line, command, kind string
  valid bool
 }{
  {"hello", "HELLO REPLY RESULT=OK VERSION=3.0\n", "HELLO", "REPLY", true},
  {"stream", "STREAM STATUS RESULT=OK\n", "STREAM", "STATUS", true},
  {"session", "SESSION STATUS RESULT=OK DESTINATION=abc\n", "SESSION", "STATUS", true},
  {"wrong prefix", "SESSION STATUS RESULT=OK\n", "STREAM", "STATUS", false},
  {"wrong kind", "STREAM REPLY RESULT=OK\n", "STREAM", "STATUS", false},
  {"missing result", "STREAM STATUS MESSAGE=hi\n", "STREAM", "STATUS", false},
  {"failure", "STREAM STATUS RESULT=I2P_ERROR\n", "STREAM", "STATUS", false},
  {"duplicate result", "STREAM STATUS RESULT=I2P_ERROR RESULT=OK\n", "STREAM", "STATUS", false},
  {"success then failure", "STREAM STATUS RESULT=OK RESULT=I2P_ERROR\n", "STREAM", "STATUS", false},
  {"misplaced success", "GARBAGE STREAM STATUS RESULT=OK\n", "STREAM", "STATUS", false},
  {"empty attribute", "STREAM STATUS RESULT=OK NAME=\n", "STREAM", "STATUS", false},
  {"unexpected token", "STREAM STATUS RESULT=OK EXTRA\n", "STREAM", "STATUS", false},
 }
 for _, tt := range tests {
  t.Run(tt.name, func(t *testing.T) {
   _, err := verifySAMReply(tt.line, tt.command, tt.kind)
   if (err == nil) != tt.valid { t.Fatalf("verifySAMReply(%q) err=%v, valid=%v", tt.line, err, tt.valid) }
  })
 }
}
