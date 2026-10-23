package rpc

import (
 "net/http"
 "net/http/httptest"
 "strings"
 "testing"
)

func TestParseSwarmIndex(t *testing.T) {
 cases := []struct { value interface{}; valid bool }{
  {float64(0), true}, {float64(2), true}, {"2", true},
  {float64(-1), false}, {"-1", false}, {float64(1.5), false},
  {"not-a-number", false}, {nil, false}, {true, false},
  {float64(9223372036854775808), false},
 }
 for _, tc := range cases {
  _, err := parseSwarmIndex(tc.value)
  if (err == nil) != tc.valid { t.Errorf("parseSwarmIndex(%v): err=%v", tc.value, err) }
 }
}

func TestRPCInvalidSwarmDoesNotPanic(t *testing.T) {
 s := NewServer(nil, "")
 for _, body := range []string{
  `{"method":"swarm-count","swarm":-1}`,
  `{"method":"swarm-count","swarm":"-1"}`,
  `{"method":"swarm-count","swarm":1.5}`,
  `{"method":"swarm-count","swarm":0}`,
 } {
  req := httptest.NewRequest(http.MethodPost, RPCPath, strings.NewReader(body))
  rec := httptest.NewRecorder()
  s.ServeHTTP(rec, req)
  if rec.Code != http.StatusOK { t.Errorf("body=%s: status=%d", body, rec.Code) }
 }
}

func TestRPCRequestBodyLimit(t *testing.T) {
 s := NewServer(nil, "")
 body := `{"method":"swarm-count","padding":"` + strings.Repeat("x", int(MaxRPCRequestBytes)) + `"}`
 rec := httptest.NewRecorder()
 s.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, RPCPath, strings.NewReader(body)))
 if rec.Code != http.StatusRequestEntityTooLarge { t.Fatalf("status=%d, want 413", rec.Code) }
}

func TestRPCMalformedJSON(t *testing.T) {
 s := NewServer(nil, "")
 rec := httptest.NewRecorder()
 s.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, RPCPath, strings.NewReader("{")))
 if rec.Code != http.StatusBadRequest { t.Fatalf("status=%d, want 400", rec.Code) }
}
