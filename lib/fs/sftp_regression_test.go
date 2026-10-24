package fs

import (
 "path"
 "strings"
 "github.com/pkg/sftp"
 "testing"
)

func TestSFTPJoinWithoutConnection(t *testing.T) {
 driver := &sftpFS{}
 got := driver.Join("/downloads", "torrent", "file.bin")
 want := path.Join("/downloads", "torrent", "file.bin")
 if got != want { t.Fatalf("Join() = %q, want %q", got, want) }
 if driver.sshClient != nil || driver.sftpClient != nil { t.Fatal("Join unexpectedly connected to server") }
}

func TestSFTPConnectionFailurePreserved(t *testing.T) {
 driver := &sftpFS{keyfile: path.Join(t.TempDir(), "missing-key")}
 err := driver.ensureConn(func(_ *sftp.Client) error { t.Fatal("callback should not run"); return nil })
 if err == nil { t.Fatal("expected missing-key connection error") }
 if !strings.Contains(err.Error(), "missing-key") { t.Fatalf("original error lost: %v", err) }
}

func TestSFTPJoinNormalizesPaths(t *testing.T) {
 driver := &sftpFS{}
 for _, tc := range []struct{ parts []string; want string }{
  {[]string{"a", "b", "c"}, "a/b/c"},
  {[]string{"/a", "b", "..", "c"}, "/a/c"},
 } {
  if got := driver.Join(tc.parts...); got != tc.want { t.Errorf("Join(%q) = %q, want %q", strings.Join(tc.parts, ","), got, tc.want) }
 }
}
