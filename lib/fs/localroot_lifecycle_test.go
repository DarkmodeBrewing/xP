package fs

import (
 "os"
 "path/filepath"
 "runtime"
 "testing"
)

func TestLocalRootRemoveTreeDoesNotFollowSymlinks(t *testing.T) {
 if runtime.GOOS == "windows" { t.Skip("symlink permissions vary on Windows") }
 parent := t.TempDir()
 inside := filepath.Join(parent, "inside")
 outside := filepath.Join(parent, "outside")
 for _, dir := range []string{inside, outside} {
  if err := os.Mkdir(dir, 0700); err != nil { t.Fatal(err) }
 }
 sentinel := filepath.Join(outside, "sentinel")
 if err := os.WriteFile(sentinel, []byte("keep"), 0600); err != nil { t.Fatal(err) }
 if err := os.Mkdir(filepath.Join(inside, "torrent"), 0700); err != nil { t.Fatal(err) }
 if err := os.WriteFile(filepath.Join(inside, "torrent", "piece"), []byte("data"), 0600); err != nil { t.Fatal(err) }
 if err := os.Symlink(outside, filepath.Join(inside, "torrent", "escape")); err != nil { t.Skipf("symlink unavailable: %v", err) }
 root, err := OpenLocalRoot(inside)
 if err != nil { t.Fatal(err) }
 defer root.Close()
 if err := root.RemoveTree("torrent"); err != nil { t.Fatal(err) }
 if _, err := os.Lstat(filepath.Join(inside, "torrent")); !os.IsNotExist(err) { t.Fatalf("torrent still exists: %v", err) }
 got, err := os.ReadFile(sentinel)
 if err != nil || string(got) != "keep" { t.Fatalf("outside data modified: %q %v", got, err) }
}

func TestLocalRootRemoveTreeRejectsTraversal(t *testing.T) {
 root, err := OpenLocalRoot(t.TempDir())
 if err != nil { t.Fatal(err) }
 defer root.Close()
 for _, name := range []string{"", ".", "../outside", "/tmp/outside"} {
  if err := root.RemoveTree(name); err == nil { t.Errorf("accepted unsafe path %q", name) }
 }
}

func TestLocalRootCopyFileToKeepsSourceAndExistingDestination(t *testing.T) {
 parent := t.TempDir()
 sourceDir := filepath.Join(parent, "source")
 targetDir := filepath.Join(parent, "target")
 for _, dir := range []string{sourceDir, targetDir} {
  if err := os.Mkdir(dir, 0700); err != nil { t.Fatal(err) }
 }
 if err := os.WriteFile(filepath.Join(sourceDir, "piece"), []byte("source"), 0600); err != nil { t.Fatal(err) }
 if err := os.WriteFile(filepath.Join(targetDir, "piece"), []byte("existing"), 0600); err != nil { t.Fatal(err) }
 src, err := OpenLocalRoot(sourceDir)
 if err != nil { t.Fatal(err) }
 defer src.Close()
 dst, err := OpenLocalRoot(targetDir)
 if err != nil { t.Fatal(err) }
 defer dst.Close()
 if err := src.CopyFileTo("piece", dst, "piece"); err == nil { t.Fatal("copy unexpectedly replaced destination") }
 for _, check := range []struct{ path, want string }{
  {filepath.Join(sourceDir, "piece"), "source"},
  {filepath.Join(targetDir, "piece"), "existing"},
 } {
  got, err := os.ReadFile(check.path)
  if err != nil || string(got) != check.want { t.Fatalf("%s: got %q, error %v", check.path, got, err) }
 }
}

func TestLocalRootCopyFileToRejectsDestinationSymlinkEscape(t *testing.T) {
 if runtime.GOOS == "windows" { t.Skip("symlink permissions vary on Windows") }
 parent := t.TempDir()
 sourceDir := filepath.Join(parent, "source")
 targetDir := filepath.Join(parent, "target")
 outside := filepath.Join(parent, "outside")
 for _, dir := range []string{sourceDir, targetDir, outside} {
  if err := os.Mkdir(dir, 0700); err != nil { t.Fatal(err) }
 }
 if err := os.WriteFile(filepath.Join(sourceDir, "piece"), []byte("data"), 0600); err != nil { t.Fatal(err) }
 if err := os.Symlink(outside, filepath.Join(targetDir, "escape")); err != nil { t.Skipf("symlink unavailable: %v", err) }
 src, err := OpenLocalRoot(sourceDir)
 if err != nil { t.Fatal(err) }
 defer src.Close()
 dst, err := OpenLocalRoot(targetDir)
 if err != nil { t.Fatal(err) }
 defer dst.Close()
 if err := src.CopyFileTo("piece", dst, "escape/piece"); err == nil { t.Fatal("copy escaped destination root") }
 if _, err := os.Stat(filepath.Join(outside, "piece")); !os.IsNotExist(err) { t.Fatalf("outside file unexpectedly exists: %v", err) }
 if got, err := os.ReadFile(filepath.Join(sourceDir, "piece")); err != nil || string(got) != "data" { t.Fatalf("source changed: %q %v", got, err) }
}

func TestLocalRootCopyFileToSuccessRetainsSourceUntilCleanup(t *testing.T) {
 parent := t.TempDir()
 sourceDir := filepath.Join(parent, "source")
 targetDir := filepath.Join(parent, "target")
 for _, dir := range []string{sourceDir, targetDir} {
  if err := os.Mkdir(dir, 0700); err != nil { t.Fatal(err) }
 }
 if err := os.WriteFile(filepath.Join(sourceDir, "piece"), []byte("payload"), 0600); err != nil { t.Fatal(err) }
 src, err := OpenLocalRoot(sourceDir)
 if err != nil { t.Fatal(err) }
 defer src.Close()
 dst, err := OpenLocalRoot(targetDir)
 if err != nil { t.Fatal(err) }
 defer dst.Close()
 if err := src.CopyFileTo("piece", dst, "nested/piece"); err != nil { t.Fatal(err) }
 for _, path := range []string{filepath.Join(sourceDir, "piece"), filepath.Join(targetDir, "nested", "piece")} {
  got, err := os.ReadFile(path)
  if err != nil || string(got) != "payload" { t.Fatalf("%s: got %q, error %v", path, got, err) }
 }
}
