# xP

**xP** is a community-maintained modernization fork of [XD](https://github.com/majestrate/XD), an I2P-focused BitTorrent client written in Go.

> **Status: reconnaissance.** This fork has not yet completed a security, performance, or compatibility audit. It is not presently claimed to be more secure or more reliable than upstream XD.

The first objective is to preserve XD's lightweight operation while establishing reproducible baselines and reviewing dependencies, security boundaries, transport interoperability, and BitTorrent protocol support. Potential future integration with **rust-mule** is exploratory, not an implemented feature.

## Upstream and attribution

xP is based on [majestrate/XD](https://github.com/majestrate/XD). Original authorship and copyright remain with the upstream contributors. See [LICENSE](LICENSE) for the MIT license. Contributions suitable for upstream will be considered for upstream submission.

## Existing functionality (inherited from XD)

- I2P-focused BitTorrent operation using SAM with i2pd or Java I2P
- Magnet URI support
- Embedded web interface (depending on build)
- CLI and RPC-related components

These are inherited capabilities, not newly audited guarantees. Actual behavior and protocol compatibility remain to be verified.

## Build and run (inherited instructions)

Prerequisites: GNU Make and a Go toolchain compatible with `go.mod` (currently Go **1.25.0**).

```sh
git clone https://github.com/DarkmodeBrewing/xP.git
cd xP
make
```

To build without the embedded web interface:

```sh
make no-webui
```

The Go module and executable are still named after upstream XD. Until any rename is separately planned and tested, follow the existing binary name:

```sh
./XD torrents.ini
```

The inherited configuration and runtime behavior are documented in [upstream XD](https://github.com/majestrate/XD#usage). Do not expose the web interface or RPC to untrusted networks without first reviewing their authentication and binding configuration.

## Reconnaissance documents

- [M0 scope and acceptance criteria](docs/modernization-roadmap.md)
- [Initial architecture inventory](docs/architecture.md)
- [Security threat-model checklist](docs/security-threat-model.md)
- [Protocol compatibility audit plan](docs/protocol-compatibility.md)
- [Performance and reliability baseline plan](docs/performance-baseline.md)

## Working principles

1. Preserve resource efficiency and compatibility; measure before optimizing.
2. Treat anonymity and network isolation as security properties requiring tests, not marketing guarantees.
3. Make small, reviewable changes; keep upstream contributions possible.
4. Audit first, then prioritize verified issues and protocol gaps.
5. Keep the core daemon independently usable; future rust-mule integration should use a versioned boundary.

## License

MIT; see [LICENSE](LICENSE). The original copyright and license notice must be retained.
