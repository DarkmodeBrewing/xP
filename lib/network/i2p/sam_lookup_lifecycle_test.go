package i2p

import (
 "errors"
 "net"
 "testing"
 "time"
)

func TestLookupUnblocksWhenSessionCloses(t *testing.T) {
 s := NewSession("test", "127.0.0.1:0", "", nil).(*samSession)
 result := make(chan error, 1)
 go func() {
  _, err := s.LookupI2P("example.i2p")
  result <- err
 }()
 select {
 case err := <-result:
  t.Fatalf("lookup unexpectedly returned before shutdown: %v", err)
 case <-time.After(20 * time.Millisecond):
 }
 if err := s.Close(); err != nil { t.Fatal(err) }
 select {
 case err := <-result:
  if !errors.Is(err, net.ErrClosed) { t.Fatalf("got %v, want net.ErrClosed", err) }
 case <-time.After(time.Second):
  t.Fatal("lookup remained blocked after session shutdown")
 }
 if err := s.Close(); err != nil { t.Fatalf("second close: %v", err) }
}

func TestLookupWorkerFailureUnblocksCallers(t *testing.T) {
 s := NewSession("test", "127.0.0.1:0", "", nil).(*samSession)
 a, b := net.Pipe()
 s.c = a
 go s.runLookups()
 result := make(chan error, 1)
 go func() {
  _, err := s.LookupI2P("example.i2p")
  result <- err
 }()
 // A broken SAM control stream should cause the worker to exit and
 // terminate the session, even if the request was already queued.
 _ = b.Close()
 select {
 case err := <-result:
  if err == nil { t.Fatal("lookup unexpectedly succeeded") }
 case <-time.After(time.Second):
  t.Fatal("lookup hung after SAM control connection failed")
 }
 select {
 case <-s.done:
 case <-time.After(time.Second):
  t.Fatal("session was not cancelled after lookup worker failed")
 }
}
