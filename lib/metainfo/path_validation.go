package metainfo

import (
 "fmt"
 "path/filepath"
 "strings"
)

// ValidatePaths rejects torrent-controlled names that could escape a storage
// directory. This is lexical validation; filesystem symlinks need separate
// driver-level protections.
func (tf *TorrentFile) ValidatePaths() error {
 if tf == nil { return fmt.Errorf("missing torrent metadata") }
 if err := validateComponent(tf.Info.Path); err != nil { return fmt.Errorf("torrent name: %w", err) }
 if !tf.IsSingleFile() {
  if len(tf.Info.Files) == 0 { return fmt.Errorf("torrent has no files") }
  for i, file := range tf.Info.Files {
   if len(file.Path) == 0 { return fmt.Errorf("file %d has empty path", i) }
   for _, component := range file.Path {
    if err := validateComponent(component); err != nil { return fmt.Errorf("file %d: %w", i, err) }
   }
  }
 }
 return nil
}

func validateComponent(component string) error {
 if component == "" || component == "." || component == ".." ||
  filepath.IsAbs(component) || strings.ContainsAny(component, "/\\\\") ||
  strings.ContainsRune(component, 0) || filepath.VolumeName(component) != "" ||
  (len(component) >= 2 && component[1] == ':' &&
   ((component[0] >= 'A' && component[0] <= 'Z') || (component[0] >= 'a' && component[0] <= 'z'))) {
  return fmt.Errorf("unsafe path component %q", component)
 }
 return nil
}
