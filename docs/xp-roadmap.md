# xP modernization roadmap

This roadmap is provisional and subject to M0 evidence. Protect the stable, efficient behavior of upstream XD.

| Phase | Focus | Exit criterion |
| --- | --- | --- |
| M0 | Inventory, baseline, threat model, test environment | Reproducible findings and prioritized backlog |
| M1 | Test coverage and CI | Unit/race/static checks and representative I2P integration smoke tests |
| M2 | Security and dependency hardening | Reviewed upgrades, regression tests for security boundaries, documented risk decisions |
| M3 | Transport and reliability | Verified lifecycle/reconnect/shutdown behavior with i2pd and Java I2P |
| M4 | Performance | Benchmarks demonstrating no unacceptable regression in idle footprint or throughput |
| M5 | Protocol modernization | Gap analysis first; implement BEP 52 / hybrid torrents only if feasible and tested |
| M6 | Integration readiness | Stable documented daemon API and lifecycle contract for possible rust-mule consumption |

## Principles

- Keep the I2P-only boundary explicit; never fall back to clearnet.
- Prefer focused, test-backed patches over broad rewrites.
- Keep upstream attribution, license notices, and Git history intact.
- Do not rename the Go module or binaries as a cosmetic documentation change; evaluate downstream imports and build/release compatibility first.
- Separate verified support from planned features, especially DHT and BitTorrent v2.
- No performance or security improvement is considered proven without measurements or tests.

## Next concrete steps

1. Execute the M0 command suite and preserve results.
2. Read the SAM/network, torrent metadata, storage, RPC, and DHT implementation paths.
3. Add a minimal CI baseline without mass dependency upgrades.
4. Open focused issues for each confirmed finding and propose changes incrementally.
