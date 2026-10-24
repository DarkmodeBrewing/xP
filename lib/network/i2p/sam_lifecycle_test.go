package i2p

import (
 "net"
 "testing"
 "time"
)

// A full lookup queue must not make session shutdown wait forever.
// The cleanup drain keeps a failing test from leaking a blocked goroutine.
func TestCloseWithFullLookupQueueReturnsPromptly(t *testing.T) {
 s := NewSession("test", "127.0.0.1:0", "", nil).(*samSession)
 a, b := net.Pipe()
 defer b.Close()
 s.c = a
 for i := 0; i < cap(s.lookup); i++ {
  s.lookup <- &lookupReq{name: "queued"}
 }
 done := make(chan error, 1)
 go func() { done <- s.Close() }()
 select {
 case <-done:
 case <-time.After(200 * time.Millisecond):
  // Release a blocked send before reporting the failure.
  <-s.lookup
  select {
  case <-done:
  case <-time.After(time.Second):
  }
  t.Fatal("Close blocked on a full lookup queue")
 }
}

// A stalled SAM peer must not block the HELLO handshake indefinitely.
// The server is closed after the observation window to release the legacy
// implementation and keep this regression test bounded.
func TestOpenControlSocketStalledHandshakeTimesOut(t *testing.T) {
 listener, err := net.Listen("tcp", "127.0.0.1:0")
 if err != nil { t.Fatal(err) }
 defer listener.Close()
 accepted := make(chan net.Conn, 1)
 go func() {
  c, err := listener.Accept()
  if err == nil { accepted <- c }
 }()
 s := NewSession("test", listener.Addr().String(), "", nil).(*samSession)
 done := make(chan error, 1)
 go func() {
  c, err := s.OpenControlSocket()
  if c != nil { _ = c.Close() }
  done <- err
 }()
 var peer net.Conn
 select {
 case peer = <-accepted:
 case <-time.After(time.Second):
  t.Fatal("SAM test server never accepted connection")
 }
 defer peer.Close()
 select {
 case err := <-done:
  if err == nil { t.Fatal("stalled SAM handshake unexpectedly succeeded") }
 case <-time.After(300 * time.Millisecond):
  _ = peer.Close()
  select {
  case <-done:
  case <-time.After(time.Second):
  }
  t.Fatal("SAM handshake has no bounded timeout")
 }
}
