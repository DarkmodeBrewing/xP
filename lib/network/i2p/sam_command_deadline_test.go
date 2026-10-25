package i2p

import (
 "bufio"
 "net"
 "testing"
 "time"
)

// A fake SAM router completes HELLO then stalls STREAM CONNECT.
// The command must time out rather than hang forever.
func TestStreamConnectStalledStatusTimesOut(t *testing.T) {
 listener, err := net.Listen("tcp", "127.0.0.1:0")
 if err != nil { t.Fatal(err) }
 defer listener.Close()
 peerDone := make(chan struct{})
 go func() {
  defer close(peerDone)
  c, e := listener.Accept()
  if e != nil { return }
  defer c.Close()
  reader := bufio.NewReader(c)
  if _, e = reader.ReadString('\n'); e != nil { return }
  if _, e = c.Write([]byte("HELLO REPLY RESULT=OK\n")); e != nil { return }
  _, _ = reader.ReadString('\n')
  // Wait until the client closes the timed-out control socket.
  var b [1]byte
  _, _ = c.Read(b[:])
 }()
 s := NewSession("test", listener.Addr().String(), "", nil).(*samSession)
 started := time.Now()
 _, err = s.DialI2P(I2PAddr("example.i2p"))
 if err == nil { t.Fatal("stalled STREAM CONNECT succeeded") }
 if time.Since(started) > samCommandTimeout+2*time.Second {
  t.Fatalf("STREAM CONNECT did not time out promptly: %v", err)
 }
 select {
 case <-peerDone:
 case <-time.After(time.Second):
  t.Fatal("timed-out connection was not closed")
 }
}

func TestSessionCreateStalledStatusTimesOut(t *testing.T) {
 s := NewSession("test", "127.0.0.1:0", "", nil).(*samSession)
 a, b := net.Pipe()
 defer b.Close()
 s.c = a
 s.keys.privkey = "test"
 done := make(chan error, 1)
 go func() { done <- s.createSession("STREAM") }()
 // Consume command so the write succeeds, but never answer.
 go func() { _, _ = bufio.NewReader(b).ReadString('\n') }()
 select {
 case err := <-done:
  if err == nil { t.Fatal("stalled SESSION CREATE succeeded") }
 case <-time.After(samCommandTimeout+2*time.Second):
  _ = b.Close()
  t.Fatal("SESSION CREATE hung beyond deadline")
 }
 _ = a.Close()
}
