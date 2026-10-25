package i2p

import (
 "fmt"
 "strings"
)

// verifySAMReply requires the exact command prefix and one unambiguous RESULT.
// Additional key/value attributes (such as VERSION or MESSAGE) are permitted.
func verifySAMReply(line, command, kind string) (map[string]string, error) {
 fields := strings.Fields(line)
 if len(fields) < 3 || fields[0] != command || fields[1] != kind {
  return nil, fmt.Errorf("invalid SAM %s %s response: %q", command, kind, line)
 }
 attrs := make(map[string]string, len(fields)-2)
 for _, field := range fields[2:] {
  key, value, ok := strings.Cut(field, "=")
  if !ok || key == "" || value == "" {
   return nil, fmt.Errorf("malformed SAM response attribute: %q", field)
  }
  if _, exists := attrs[key]; exists {
   return nil, fmt.Errorf("duplicate SAM response attribute: %q", key)
  }
  attrs[key] = value
 }
 result, ok := attrs["RESULT"]
 if !ok { return nil, fmt.Errorf("SAM response missing RESULT: %q", line) }
 if result != "OK" { return nil, fmt.Errorf("SAM command failed (%s): %q", result, line) }
 return attrs, nil
}

 
// verifySAMFields validates a response without requiring RESULT (DEST REPLY).
func verifySAMFields(line, command, kind string) (map[string]string, error) {
 fields := strings.Fields(line)
 if len(fields) < 3 || fields[0] != command || fields[1] != kind {
  return nil, fmt.Errorf("invalid SAM %s %s response: %q", command, kind, line)
 }
 attrs := make(map[string]string, len(fields)-2)
 for _, field := range fields[2:] {
  key, value, ok := strings.Cut(field, "=")
  if !ok || key == "" || value == "" {
   return nil, fmt.Errorf("malformed SAM response attribute: %q", field)
  }
  if _, exists := attrs[key]; exists {
   return nil, fmt.Errorf("duplicate SAM response attribute: %q", key)
  }
  attrs[key] = value
 }
 return attrs, nil
}
