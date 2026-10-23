package metainfo

import "testing"

func TestValidatePaths(t *testing.T) {
 tests := []struct {
  name string
  root string
  files []FilePath
  invalid bool
 }{
  {"normal", "album", []FilePath{{"disc 1", "song.flac"}}, false},
  {"unicode", "musik", []FilePath{{"låtar", "sång.flac"}}, false},
  {"parent root", "..", []FilePath{{"song.flac"}}, true},
  {"absolute root", "/tmp", []FilePath{{"song.flac"}}, true},
  {"slash root", "a/b", []FilePath{{"song.flac"}}, true},
  {"traversal", "album", []FilePath{{"..", "escape"}}, true},
  {"embedded traversal", "album", []FilePath{{"../escape"}}, true},
  {"absolute file", "album", []FilePath{{"/etc/passwd"}}, true},
  {"backslash", "album", []FilePath{{"..\\escape"}}, true},
  {"drive", "album", []FilePath{{"C:\\secret"}}, true},
  {"empty file", "album", []FilePath{{}}, true},
  {"empty root", "", []FilePath{{"song.flac"}}, true},
  {"dot", "album", []FilePath{{"."}}, true},
  {"nul", "album", []FilePath{{"bad\x00name"}}, true},
 }
 for _, tc := range tests {
  t.Run(tc.name, func(t *testing.T) {
   info := Info{Path:tc.root}
   for _, p := range tc.files { info.Files = append(info.Files, FileInfo{Path:p, Length:1}) }
   tf := &TorrentFile{Info:info}
   err := tf.ValidatePaths()
   if (err != nil) != tc.invalid { t.Fatalf("ValidatePaths() err=%v, invalid=%v",err,tc.invalid) }
  })
 }
}

func TestValidateSingleFileName(t *testing.T) {
 for _, name := range []string{"../escape", "/tmp/escape", "..", "C:\\escape"} {
  tf := &TorrentFile{Info:Info{Path:name,Length:1}}
  if err:=tf.ValidatePaths(); err==nil { t.Errorf("accepted %q",name) }
 }
 tf := &TorrentFile{Info:Info{Path:"good.txt",Length:1}}
 if err:=tf.ValidatePaths(); err!=nil { t.Fatal(err) }
}
