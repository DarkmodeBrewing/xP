package fs

import (
 "os"
 "path/filepath"
 "runtime"
 "testing"
)

func TestLocalRootRejectsSymlinkEscape(t *testing.T) {
 if runtime.GOOS=="windows" { t.Skip("symlink privileges vary on Windows") }
 parent:=t.TempDir()
 inside:=filepath.Join(parent,"inside")
 outside:=filepath.Join(parent,"outside")
 if err:=os.Mkdir(inside,0700);err!=nil {t.Fatal(err)}
 if err:=os.Mkdir(outside,0700);err!=nil {t.Fatal(err)}
 if err:=os.WriteFile(filepath.Join(outside,"secret"),[]byte("untouched"),0600);err!=nil {t.Fatal(err)}
 if err:=os.Symlink(outside,filepath.Join(inside,"link"));err!=nil {t.Skipf("symlinks unavailable: %v",err)}
 r,err:=OpenLocalRoot(inside);if err!=nil {t.Fatal(err)}
 defer r.Close()
 if f,err:=r.OpenRead("link/secret");err==nil {f.Close();t.Fatal("read escaped root")}
 if f,err:=r.OpenWrite("link/new");err==nil {f.Close();t.Fatal("write escaped root")}
 if err:=r.EnsureFile("link/created",1);err==nil {t.Fatal("allocation escaped root")}
 if err:=r.Remove("link/secret");err==nil {t.Fatal("remove escaped root")}
 if _,err:=os.Stat(filepath.Join(outside,"new"));!os.IsNotExist(err) {t.Fatalf("outside changed: %v",err)}
 if b,err:=os.ReadFile(filepath.Join(outside,"secret"));err!=nil||string(b)!="untouched" {t.Fatalf("outside secret changed: %q %v",b,err)}
}

func TestLocalRootValidNestedFile(t *testing.T) {
 root,err:=OpenLocalRoot(t.TempDir());if err!=nil {t.Fatal(err)}
 defer root.Close()
 if err:=root.EnsureFile("music/disc/song",8);err!=nil {t.Fatal(err)}
 f,err:=root.OpenWrite("music/disc/song");if err!=nil {t.Fatal(err)}
 if _,err=f.WriteAt([]byte("test"),0);err!=nil {t.Fatal(err)}
 f.Close()
 f,err=root.OpenRead("music/disc/song");if err!=nil {t.Fatal(err)}
 defer f.Close()
 buf:=make([]byte,4)
 if _,err=f.ReadAt(buf,0);err!=nil {t.Fatal(err)}
 if string(buf)!="test" {t.Fatalf("got %q",buf)}
}

func TestLocalRootRejectsRelativeTraversal(t *testing.T) {
 root,err:=OpenLocalRoot(t.TempDir());if err!=nil {t.Fatal(err)}
 defer root.Close()
 for _,p:=range []string{"../outside","/etc/passwd","","."} {
  if f,err:=root.OpenWrite(p);err==nil {f.Close();t.Errorf("accepted %q",p)}
 }
}
