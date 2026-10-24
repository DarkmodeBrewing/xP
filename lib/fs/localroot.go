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
