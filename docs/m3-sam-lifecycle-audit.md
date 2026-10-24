# M3.1 — SAM transport lifecycle and failure-mode audit

Status: **audit plan, not a verified implementation review**.

## Invariants

1. xP must never fall back to clearnet when SAM or the I2P router is unavailable.
2. A failed session must not be advertised as connected.
3. Shutdown must stop accepting new work and close owned sessions, streams, listeners, and goroutines.
4. Reconnection must not create duplicate sessions or leak file descriptors and goroutines.
5. Retry loops must be bounded by backoff and respect shutdown cancellation.
6. Router-specific behavior must be checked against both i2pd and Java I2P.

## Inventory to collect

Trace the call chain from daemon startup and configuration through SAM dialing, handshake, destination/session creation, stream accept/dial, torrent peer connections, shutdown, and retry logic. Record the exact owning type, goroutine, close path, timeout, and error propagation for every connection.

Document which endpoints are configured and whether any ordinary TCP dial can bypass the SAM boundary.

## Failure matrix

| Scenario | Required behavior | Verification |
| --- | --- | --- |
| Router unavailable at startup | Return clear error or bounded retry; no clearnet | Offline integration test |
| SAM handshake fails | Close connection; preserve original error | Fake SAM peer |
| Session creation fails | No leaked socket or half-initialized session | Failure injection |
| Router disconnects while idle | Detect failure; transition to disconnected | Router restart test |
| Router disconnects during transfer | Fail affected streams; recover without corrupting state | Active-transfer test |
| Repeated reconnect attempts | Backoff; no duplicate sessions or goroutine growth | Repeated-cycle test |
| Shutdown during dial/backoff | Cancel promptly and release resources | Cancellation test |
| Router returns | Re-establish session and resume normal peer handling | i2pd / Java I2P smoke test |

## Proposed first implementation slice

1. Identify the actual SAM connection and lifecycle entry points before changing code.
2. Add tests around the existing implementation, starting with router-unavailable startup and shutdown cancellation.
3. Introduce explicit lifecycle state or cancellation only where observed behavior requires it.
4. Measure retries, resource counts, and recovery time; do not claim improvement without results.

## Exit criteria

- Code-path inventory cites actual symbols and files.
- Offline failure and shutdown tests run in CI.
- Router restart smoke tests documented for i2pd and Java I2P.
- No-clearnet invariant tested or explicitly marked as an unresolved security risk.
- Remaining issues prioritized with reproductions and owners.

This document does not claim that the transport implementation already satisfies these invariants.
