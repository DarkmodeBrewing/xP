# xP M0 — Codebase reconnaissance

Status: **initial inventory only; tests and audits not yet executed** (2026-10-23).

## Goals and guardrails

Preserve XD's low resource usage, reliability, and I2P-only network isolation. Make changes evidence-led and suitable for upstream contribution where possible. No automatic clearnet fallback. Do not change the Go module path or executable name before a compatibility assessment.

## Observed repository inventory

- Default branch: `master`; fork of `majestrate/XD`.
- Go module: `github.com/majestrate/XD`; `go.mod` specifies Go 1.25.0. Existing README says Go 1.24+, requiring correction after build verification.
- Direct dependencies currently declared: `github.com/pkg/sftp v1.13.5`, `github.com/zeebo/bencode v1.0.0`, `golang.org/x/crypto v0.52.0`, `gopkg.in/leonelquinteros/gotext.v1 v1.3.1`.
- Top-level directories include `cmd`, `lib`, `docs`, `contrib`, and `debian`.
- `lib` package names include `bittorrent`, `dht`, `metainfo`, `network`, `rpc`, `storage`, `tracker`, `fs`, `sync`, `config`, and `gnutella`. Package presence does **not** establish completed or enabled protocol support.
- README describes SAM support for i2pd and Java I2P, magnets, and an embedded web UI. Claims require runtime validation.

## Reproducible baseline commands

Run in an isolated development environment with a supported Go toolchain:

```sh
go version
go mod download
go list -m all
go test ./...
go test -race ./...
go vet ./...
go build ./...
govulncheck ./...   # install golang.org/x/vuln/cmd/govulncheck separately
```

Record toolchain version, platform, exact commit SHA, command output, failures, and resource usage. Never mark these checks passing until run. Add `go test -fuzz` targets only after locating suitable parsers and harnesses.

## Audit tracks

1. **Security:** torrent/metainfo parsing limits, malformed bencode, path traversal and symlink handling, untrusted filenames, RPC/web UI bind/auth/CSRF, resource exhaustion, dependency vulnerabilities, and I2P-only egress invariants.
2. **Transport:** SAM version negotiation, destination lifecycle, session shutdown, reconnect/backoff, timeouts, router restart, and i2pd/Java I2P compatibility.
3. **Stability:** concurrency races, goroutine/file descriptor leaks, resume-state integrity, incomplete downloads, cancellation and clean shutdown.
4. **Performance:** idle and active CPU/RSS, goroutines, allocation profiles, disk I/O, peer counts, and sustained throughput, using identical workloads before/after changes.
5. **Protocol:** inspect existing BEP support before proposing BEP 52 (v2 metadata, SHA-256 Merkle trees, piece layers, hybrid torrents, magnet identifiers) and I2P-specific interoperability.
6. **Supply chain:** SBOM, module provenance, dependency freshness, build/release automation, licenses and artifact integrity.

## Acceptance for M0 completion

- Baseline commands executed with attached logs and failures triaged.
- Architecture and trust boundaries mapped to concrete code paths.
- Dependency and vulnerability inventory captured.
- Transport/protocol capability matrix verified, not inferred from filenames.
- Prioritized findings with severity, reproduction steps, and test proposals.

## Current limitations

This document is a plan plus repository metadata reconnaissance, **not** a completed code audit. No vulnerability, performance, or compatibility claim has been established by testing.
