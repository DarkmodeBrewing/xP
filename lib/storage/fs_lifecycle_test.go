package storage

import (
 "os"
 "path/filepath"
 "testing"

 "github.com/majestrate/XD/lib/fs"
 "github.com/majestrate/XD/lib/metainfo"
)

func fixtureTorrent(t *testing.T, dir string) *fsTorrent {
 t.Helper()
 info, err := metainfo.TorrentFileFromInfo(metainfo.Info{
  Path: "bundle",
  Files: []metainfo.FileInfo{
   {Path: metainfo.FilePath{"one"}, Length: 3},
   {Path: metainfo.FilePath{"two"}, Length: 3},
  },
 })
 if err != nil { t.Fatal(err) }
 return &fsTorrent{dir: dir, meta: info, st: &FsStorage{FS: fs.STD, MetaDir: t.TempDir()}}
}

func TestLocalTorrentMoveFailurePreservesOriginal(t *testing.T) {
 parent := t.TempDir()
 source := filepath.Join(parent, "source")
 target := filepath.Join(parent, "target")
 for _, dir := range []string{source, target, filepath.Join(source, "bundle"), filepath.Join(target, "bundle")} {
  if err := os.Mkdir(dir, 0700); err != nil { t.Fatal(err) }
 }
 for _, name := range []string{"one", "two"} {
  if err := os.WriteFile(filepath.Join(source, "bundle", name), []byte(name), 0600); err != nil { t.Fatal(err) }
 }
 // A collision on the second file forces rollback of the first copy.
 if err := os.WriteFile(filepath.Join(target, "bundle", "two"), []byte("existing"), 0600); err != nil { t.Fatal(err) }
 torrent := fixtureTorrent(t, source)
 if err := torrent.MoveTo(target); err == nil { t.Fatal("move unexpectedly succeeded") }
 if torrent.dir != source { t.Fatal("torrent directory changed after failed move") }
 for _, name := range []string{"one", "two"} {
  got, err := os.ReadFile(filepath.Join(source, "bundle", name))
  if err != nil || string(got) != name { t.Fatalf("source %s: %q %v", name, got, err) }
 }
 if _, err := os.Stat(filepath.Join(target, "bundle", "one")); !os.IsNotExist(err) { t.Fatalf("partial copy not rolled back: %v", err) }
 got, err := os.ReadFile(filepath.Join(target, "bundle", "two"))
 if err != nil || string(got) != "existing" { t.Fatalf("existing target changed: %q %v", got, err) }
}

func TestLocalTorrentMoveAndDeleteLifecycle(t *testing.T) {
 parent := t.TempDir()
 source := filepath.Join(parent, "source")
 target := filepath.Join(parent, "target")
 if err := os.Mkdir(source, 0700); err != nil { t.Fatal(err) }
 if err := os.Mkdir(target, 0700); err != nil { t.Fatal(err) }
 torrent := fixtureTorrent(t, source)
 if err := torrent.Allocate(); err != nil { t.Fatal(err) }
 for _, name := range []string{"one", "two"} {
  if err := os.WriteFile(filepath.Join(source, "bundle", name), []byte(name), 0600); err != nil { t.Fatal(err) }
 }
 if err := torrent.MoveTo(target); err != nil { t.Fatal(err) }
 if torrent.dir != target { t.Fatalf("unexpected torrent directory %q", torrent.dir) }
 if _, err := os.Stat(filepath.Join(source, "bundle")); !os.IsNotExist(err) { t.Fatalf("source tree remains: %v", err) }
 for _, name := range []string{"one", "two"} {
  got, err := os.ReadFile(filepath.Join(target, "bundle", name))
  if err != nil || string(got) != name { t.Fatalf("target %s: %q %v", name, got, err) }
 }
 if err := torrent.Delete(); err != nil { t.Fatal(err) }
 if _, err := os.Stat(filepath.Join(target, "bundle")); !os.IsNotExist(err) { t.Fatalf("deleted torrent remains: %v", err) }
}
