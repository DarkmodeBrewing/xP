package fs

import (
 "net"
 "os"
 "path/filepath"
 "testing"

 "github.com/pkg/sftp"
)

// testSFTPServer uses a real SFTP protocol server over an in-memory transport.
// All filesystem paths are confined to t.TempDir by the test fixtures.
func testSFTPServer(t *testing.T) *sftpFS {
 t.Helper()
 clientConn, serverConn := net.Pipe()
 server, err := sftp.NewServer(serverConn)
 if err != nil { t.Fatal(err) }
 done := make(chan error, 1)
 go func() { done <- server.Serve() }()
 client, err := sftp.NewClientPipe(clientConn, clientConn)
 if err != nil {
  _ = clientConn.Close()
  _ = server.Close()
  t.Fatal(err)
 }
 t.Cleanup(func() {
  _ = client.Close()
  _ = clientConn.Close()
  _ = server.Close()
  <-done
 })
 return &sftpFS{sftpClient: client}
}

func TestSFTPRemoveAllDoesNotFollowNestedSymlink(t *testing.T) {
 root := t.TempDir()
 outside := filepath.Join(root, "outside")
 torrent := filepath.Join(root, "torrent")
 if err := os.Mkdir(outside, 0700); err != nil { t.Fatal(err) }
 if err := os.Mkdir(torrent, 0700); err != nil { t.Fatal(err) }
 sentinel := filepath.Join(outside, "keep")
 if err := os.WriteFile(sentinel, []byte("safe"), 0600); err != nil { t.Fatal(err) }
 if err := os.WriteFile(filepath.Join(torrent, "data"), []byte("torrent"), 0600); err != nil { t.Fatal(err) }
 if err := os.Symlink(outside, filepath.Join(torrent, "link")); err != nil { t.Skipf("symlinks unsupported: %v", err) }
 driver := testSFTPServer(t)
 if err := driver.RemoveAll(torrent); err != nil { t.Fatal(err) }
 if _, err := os.Stat(torrent); !os.IsNotExist(err) { t.Fatalf("torrent still exists: %v", err) }
 if got, err := os.ReadFile(sentinel); err != nil || string(got) != "safe" { t.Fatalf("symlink target damaged: %q %v", got, err) }
}

func TestSFTPRemoveAllTopLevelSymlink(t *testing.T) {
 root := t.TempDir()
 outside := filepath.Join(root, "outside")
 link := filepath.Join(root, "link")
 if err := os.Mkdir(outside, 0700); err != nil { t.Fatal(err) }
 sentinel := filepath.Join(outside, "keep")
 if err := os.WriteFile(sentinel, []byte("safe"), 0600); err != nil { t.Fatal(err) }
 if err := os.Symlink(outside, link); err != nil { t.Skipf("symlinks unsupported: %v", err) }
 driver := testSFTPServer(t)
 if err := driver.RemoveAll(link); err != nil { t.Fatal(err) }
 if _, err := os.Lstat(link); !os.IsNotExist(err) { t.Fatalf("symlink still exists: %v", err) }
 if got, err := os.ReadFile(sentinel); err != nil || string(got) != "safe" { t.Fatalf("target damaged: %q %v", got, err) }
}

func TestSFTPRemoveAllNestedDirectories(t *testing.T) {
 root := t.TempDir()
 torrent := filepath.Join(root, "torrent")
 nested := filepath.Join(torrent, "one", "two")
 if err := os.MkdirAll(nested, 0700); err != nil { t.Fatal(err) }
 if err := os.WriteFile(filepath.Join(nested, "payload"), []byte("data"), 0600); err != nil { t.Fatal(err) }
 driver := testSFTPServer(t)
 if err := driver.RemoveAll(torrent); err != nil { t.Fatal(err) }
 if _, err := os.Stat(torrent); !os.IsNotExist(err) { t.Fatalf("nested tree remains: %v", err) }
}
