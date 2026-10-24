package fs

import (
 "fmt"
 "io"
 "os"
 "path/filepath"
)

// LocalRoot confines file operations to an already configured local storage
// directory. Paths are relative to the root; os.Root prevents symlink escapes
// and parent-directory replacement races during each operation.
type LocalRoot struct { root *os.Root }

func OpenLocalRoot(dir string) (*LocalRoot, error) {
 r, err := os.OpenRoot(dir)
 if err != nil { return nil, err }
 return &LocalRoot{root:r}, nil
}

func (r *LocalRoot) Close() error { return r.root.Close() }

func validateRelativePath(name string) error {
 if name == "" || name == "." || filepath.IsAbs(name) || !filepath.IsLocal(name) {
  return fmt.Errorf("unsafe relative path %q", name)
 }
 return nil
}

func (r *LocalRoot) OpenRead(name string) (*os.File, error) {
 if err:=validateRelativePath(name); err!=nil { return nil,err }
 return r.root.Open(name)
}

func (r *LocalRoot) OpenWrite(name string) (*os.File, error) {
 if err:=validateRelativePath(name); err!=nil { return nil,err }
 return r.root.OpenFile(name, os.O_WRONLY|os.O_CREATE, 0666)
}

func (r *LocalRoot) EnsureFile(name string, size uint64) error {
 if err:=validateRelativePath(name); err!=nil { return err }
 if size > uint64(1<<63-1) { return fmt.Errorf("file size too large") }
 if err:=r.root.MkdirAll(filepath.Dir(name), 0755); err!=nil { return err }
 f,err:=r.root.OpenFile(name,os.O_WRONLY|os.O_CREATE|os.O_EXCL,0666)
 if os.IsExist(err) { return nil }
 if err!=nil { return err }
 defer f.Close()
 if size > 0 {
  if err=f.Truncate(int64(size)); err!=nil { return err }
 }
 return nil
}

func (r *LocalRoot) Remove(name string) error {
 if err:=validateRelativePath(name); err!=nil { return err }
 return r.root.Remove(name)
}

// CopyTo writes to a root-relative path without resolving a destination
// symlink outside the root. Callers must close the source.
func (r *LocalRoot) CopyTo(name string, src io.Reader) (int64,error) {
 if err:=validateRelativePath(name); err!=nil { return 0,err }
 if err:=r.root.MkdirAll(filepath.Dir(name),0755); err!=nil { return 0,err }
 f,err:=r.root.OpenFile(name,os.O_CREATE|os.O_WRONLY|os.O_TRUNC,0666)
 if err!=nil { return 0,err }
 n,err:=io.Copy(f,src)
 closeErr:=f.Close()
 if err!=nil { return n,err }
 return n,closeErr
}

// RemoveTree removes a relative file or directory without following symlinks.
// The root descriptor confines each operation to its configured directory.
func (r *LocalRoot) RemoveTree(name string) error {
 if err := validateRelativePath(name); err != nil { return err }
 info, err := r.root.Lstat(name)
 if os.IsNotExist(err) { return nil }
 if err != nil { return err }
 if !info.IsDir() { return r.root.Remove(name) }
 dir, err := r.root.Open(name)
 if err != nil { return err }
 entries, readErr := dir.ReadDir(-1)
 closeErr := dir.Close()
 if readErr != nil { return readErr }
 if closeErr != nil { return closeErr }
 for _, entry := range entries {
  if err := r.RemoveTree(filepath.Join(name, entry.Name())); err != nil { return err }
 }
 return r.root.Remove(name)
}

// CopyFileTo copies a regular file into another root without replacing an
// existing destination. The caller owns source cleanup after all copies succeed.
func (r *LocalRoot) CopyFileTo(name string, dst *LocalRoot, target string) error {
 if err := validateRelativePath(name); err != nil { return err }
 if err := validateRelativePath(target); err != nil { return err }
 src, err := r.OpenRead(name)
 if err != nil { return err }
 defer src.Close()
 info, err := src.Stat()
 if err != nil { return err }
 if !info.Mode().IsRegular() { return fmt.Errorf("source %q is not regular", name) }
 if err := dst.root.MkdirAll(filepath.Dir(target), 0755); err != nil { return err }
 out, err := dst.root.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
 if err != nil { return err }
 _, copyErr := io.Copy(out, src)
 if copyErr == nil { copyErr = out.Sync() }
 closeErr := out.Close()
 if copyErr != nil || closeErr != nil {
  _ = dst.root.Remove(target)
  if copyErr != nil { return copyErr }
  return closeErr
 }
 return nil
}
