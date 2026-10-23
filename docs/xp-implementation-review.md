# xP — M1 initial implementation review

Date: 2026-10-23. Status: **static inspection; not an exploit confirmation or completed audit**.

## Evidence-backed observations

### I2P SAM transport

- `lib/network/i2p/session.go`: `NewSession` configures `minversion` and `maxversion` as `3.0`. Investigate feature negotiation and current router compatibility. Do not bump versions blindly.
- `lib/network/i2p/sam.go`: session setup, naming lookup and shutdown behavior merit integration tests with i2pd and Java I2P. Review blocked lookups on control socket failure and lifecycle behavior.
- `lib/network/inet/inet.go`: separate inet/Lokinet-oriented implementation exists. Audit build tags and runtime selection before claiming all possible builds are I2P-only.

### Torrent metadata and BitTorrent v2

- `lib/metainfo/metainfo.go`: `Infohash` uses SHA-1 of raw info bytes; `CheckPiece` compares SHA-1 against 20-byte hashes. These are v1 structures. No v2 support has been verified.
- `CheckPiece` slices `i.Pieces[idx:idx+20]` after comparing against `NumPieces()`. Test malformed metadata, integer boundaries, and whether piece index multiplication can overflow before adding validation.
- `BDecode` decodes untrusted bencoded metadata from an `io.Reader`; trace callers to establish input limits and validation before allocating files.

### Filesystem and storage

- `lib/metainfo/metainfo.go`: torrent-controlled `name` and `path` components are joined with `filepath.Join`.
- `lib/storage/fs.go`: file allocation, reads, writes, moves, and deletion use derived paths. Review `lib/fs` confinement, path component validation, symlink policy and cross-platform behavior. No traversal exploit is claimed.
- `readFileAt` obtains a file handle then uses it; investigate whether error paths are safe and add regression coverage.

### HTTP RPC

- `lib/rpc/server.go`: checks `req.Host` against an expected host or localhost, but this is not authentication. Determine actual listen address and exposure from configuration.
- `ServeHTTP` decodes JSON from `req.Body` without a visible per-request byte limit in this handler. Check outer HTTP server limits and test oversized payloads.
- The parsed `swarmidx` is compared with `len(r.sw)` but not visibly checked for negativity before indexing. Add a negative-index regression test and inspect the full call path.
- `NewServer` accesses `sw[0]`; validate empty-swarm construction behavior.
- `lib/rpc/transmission/handler_notimpl.go` confirms some Transmission-compatible methods have a not-implemented response; inventory actual implemented methods before advertising compatibility.

### Piece handling

- `lib/bittorrent/swarm/piece.go`: `cachedPiece.accept` checks `offset+length > p.length`; review unsigned overflow and zero-length handling. `finalChunkBitfieldIndex` uses `p.length-1`; check zero-length preconditions.

## Prioritized follow-ups

| Priority | Candidate | Validation |
| --- | --- | --- |
| P0 | RPC negative swarm index | Unit test with negative input, expect 4xx/error without panic |
| P0 | Filesystem path confinement | Malicious torrent path tests and symlink escape tests |
| P1 | Unbounded RPC input | Inspect server limits and add oversized-body test |
| P1 | Malformed metadata/piece bounds | Table tests and fuzz targets |
| P1 | SAM lifecycle | Router restart, failed lookup, shutdown/reconnect integration tests |
| P2 | BEP 52 | Capability matrix and design before implementation |
| P2 | Performance | Baseline CPU/RSS/allocations and repeatable transfer workloads |

## CI introduced in this slice

`.github/workflows/go-ci.yml` runs package tests, `go vet`, compilation, race tests, and `govulncheck` on GitHub-hosted Ubuntu. CI results are unknown until GitHub Actions runs. Tests do not prove transport interoperability or security.

## Next iteration

First reproduce and test the RPC indexing issue, then examine `lib/fs` and torrent ingestion boundaries. Keep implementation fixes in separate focused PRs so regression risk remains measurable.
